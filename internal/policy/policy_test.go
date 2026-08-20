package policy

import "testing"

// TestPolicyReviewStateValidR020A review 必须是合法结果状态。
func TestPolicyReviewStateValidR020A(t *testing.T) {
	if !ValidReviewState(ReviewState) {
		t.Fatal("review should be a valid result state")
	}
	if ReviewStateLabel(ReviewState) == "" {
		t.Fatal("review should have a label")
	}
}

// TestPolicyReviewTransitionR020B review 中间态必须能流转到 pass/fail。
func TestPolicyReviewTransitionR020B(t *testing.T) {
	ok, err := CanReviewTransition(ReviewState, "pass")
	if err != nil || !ok {
		t.Fatalf("review -> pass should be allowed, ok=%v err=%v", ok, err)
	}
	ok, err = CanReviewTransition(ReviewState, "fail")
	if err != nil || !ok {
		t.Fatalf("review -> fail should be allowed, ok=%v err=%v", ok, err)
	}
	ok, err = CanReviewTransition("pending", ReviewState)
	if err != nil || !ok {
		t.Fatalf("pending -> review should be allowed, ok=%v err=%v", ok, err)
	}
}

// TestPolicyReleaseBlocksReviewR020C 复核中的批次不能放行。
func TestPolicyReleaseBlocksReviewR020C(t *testing.T) {
	if err := BatchReleaseRule(10, false, true); err == nil {
		t.Fatal("expected a batch under review to be blocked from release")
	}
}

// TestPolicyRecallBlocksReviewR020D 复核中的批次不能召回。
func TestPolicyRecallBlocksReviewR020D(t *testing.T) {
	if _, err := RecallLevelRule(90, true); err == nil {
		t.Fatal("expected a batch under review to be blocked from recall")
	}
}

// TestPolicyRxBlocksReviewR020F 复核中的批次暂停处方药销售。
func TestPolicyRxBlocksReviewR020F(t *testing.T) {
	if err := RxSaleRule("rx", true, true, true); err == nil {
		t.Fatal("expected rx sale to be suspended while under review")
	}
}
