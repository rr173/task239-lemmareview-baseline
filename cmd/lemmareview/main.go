// Command lemmareview 是数学证明引理依赖复核台的服务入口。
// 支持 --addr（监听地址）、--db（SQLite 路径）、--smoke-test（自检并退出）。
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"task239-lemmareview/internal/httpapi"
	"task239-lemmareview/internal/model"
	"task239-lemmareview/internal/service"
	"task239-lemmareview/internal/store"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	dbPath := flag.String("db", "lemmareview.db", "SQLite database path")
	smoke := flag.Bool("smoke-test", false, "run self-check then exit (no server)")
	flag.Parse()

	if *smoke {
		if err := runSmoke(*dbPath); err != nil {
			log.Fatalf("smoke-test failed: %v", err)
		}
		fmt.Println("smoke-test passed")
		os.Exit(0)
	}

	st, err := store.New(*dbPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer st.Close()

	svc := service.New(st)
	srv := httpapi.New(svc, *addr, *dbPath)
	if err := srv.Run(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// runSmoke 真实创建草稿、写入步骤与引理、构建前提依赖、执行覆盖分析、
// 检测循环、冻结版本，关闭并重新打开数据库验证持久化与重启恢复。
func runSmoke(dbPath string) error {
	_ = os.Remove(dbPath)
	_ = os.Remove(dbPath + "-wal")
	_ = os.Remove(dbPath + "-shm")

	st, err := store.New(dbPath)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	svc := service.New(st)

	// 1. 创建草稿
	draft, err := svc.CreateDraft("smoke-draft", "self-check draft")
	if err != nil {
		return fmt.Errorf("create draft: %w", err)
	}

	// 2. 导入步骤：含一个间接自循环（步骤 3 依赖步骤 2，步骤 2 依赖步骤 3 → 循环）
	steps, err := svc.ImportSteps(draft.ID, `1 已知前提 A => A成立
2 由 A 推出 B => B成立
3 由 B 与 C 推出结论 D => D成立
4 由 D 推出最终定理 => 定理成立`)
	if err != nil {
		return fmt.Errorf("import steps: %w", err)
	}

	// 3. 创建引理 C（步骤 3 依赖）
	lemmaC, err := svc.CreateLemma(draft.ID, "引理C", "C 成立的条件")
	if err != nil {
		return fmt.Errorf("create lemma: %w", err)
	}
	if err := svc.Store().UpdateLemmaStatus(lemmaC.ID, model.LemmaAvailable); err != nil {
		return fmt.Errorf("mark lemma available: %w", err)
	}

	// 4. 添加前提边：步骤3 依赖 引理C
	if err := svc.AddPremise(draft.ID, lemmaC.ID, "lemma", steps[2].ID, true); err != nil {
		return fmt.Errorf("add premise lemma: %w", err)
	}
	// 步骤2 依赖 步骤1（顺序合法）
	if err := svc.AddPremise(draft.ID, steps[0].ID, "step", steps[1].ID, true); err != nil {
		return fmt.Errorf("add premise step: %w", err)
	}
	// 步骤3 依赖 步骤2
	if err := svc.AddPremise(draft.ID, steps[1].ID, "step", steps[2].ID, true); err != nil {
		return fmt.Errorf("add premise step2: %w", err)
	}
	// 步骤4 依赖 步骤3
	if err := svc.AddPremise(draft.ID, steps[2].ID, "step", steps[3].ID, true); err != nil {
		return fmt.Errorf("add premise step3: %w", err)
	}

	// 5. 覆盖分析
	res, err := svc.Analyze(draft.ID)
	if err != nil {
		return fmt.Errorf("analyze: %w", err)
	}
	if len(res.CoveredSteps) == 0 {
		return fmt.Errorf("expected covered steps, got none")
	}

	// 6. 冻结版本
	v, err := svc.FreezeVersion(draft.ID, "smoke-v1")
	if err != nil {
		return fmt.Errorf("freeze: %w", err)
	}
	if v.Status != model.VersionFrozen {
		return fmt.Errorf("version not frozen: %s", v.Status)
	}

	// 7. 关闭并重新打开数据库，验证持久化与重启恢复
	if err := st.Close(); err != nil {
		return fmt.Errorf("close: %w", err)
	}
	st2, err := store.New(dbPath)
	if err != nil {
		return fmt.Errorf("reopen: %w", err)
	}
	defer st2.Close()
	svc2 := service.New(st2)

	d2, err := svc2.GetDraft(draft.ID)
	if err != nil {
		return fmt.Errorf("reopen get draft: %w", err)
	}
	if d2.Status != model.DraftFrozen {
		return fmt.Errorf("reopen draft status mismatch: %s", d2.Status)
	}
	steps2, err := svc2.ListSteps(draft.ID)
	if err != nil {
		return fmt.Errorf("reopen list steps: %w", err)
	}
	if len(steps2) != 4 {
		return fmt.Errorf("reopen steps count mismatch: %d", len(steps2))
	}
	vs, err := svc2.ListVersions(draft.ID)
	if err != nil {
		return fmt.Errorf("reopen list versions: %w", err)
	}
	if len(vs) != 1 {
		return fmt.Errorf("reopen versions count mismatch: %d", len(vs))
	}

	// 8. 冻结草稿应拒绝修改
	if _, err := svc2.ImportSteps(draft.ID, "5 新步骤 => x"); err == nil {
		return fmt.Errorf("expected frozen write rejection")
	}
	return nil
}
