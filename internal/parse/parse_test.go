package parse

import "testing"

func TestParseStepsPreservesLabelsAndConclusion(t *testing.T) {
	steps, err := ParseSteps("2.1 derive relation => relation holds\n2.2 finish")
	if err != nil {
		t.Fatal(err)
	}
	if len(steps) != 2 || steps[0].Label != "2.1" || steps[0].Conclusion != "relation holds" {
		t.Fatalf("unexpected parsed steps: %#v", steps)
	}
	if steps[1].Seq != 2 || steps[1].Statement != "finish" {
		t.Fatalf("unexpected second step: %#v", steps[1])
	}
}

func TestValidateLabelsUniqueRejectsDuplicate(t *testing.T) {
	steps, err := ParseSteps("1 first\n1 repeated")
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateLabelsUnique(steps); err == nil {
		t.Fatal("expected duplicate label error")
	}
}
