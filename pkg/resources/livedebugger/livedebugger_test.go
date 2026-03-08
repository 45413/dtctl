package livedebugger

import "testing"

func TestBuildFilterSets_UsesLabelsAndEmptyFilters(t *testing.T) {
	input := map[string][]string{
		"k8s.container.name":          {"credit-card-order-service"},
		"dt.kubernetes.workload.name": {"credit-card-order-service"},
	}

	sets := BuildFilterSets(input)
	if len(sets) != 1 {
		t.Fatalf("expected one filter set, got %d", len(sets))
	}

	set := sets[0]

	labels, ok := set["labels"].([]map[string]interface{})
	if !ok {
		t.Fatalf("expected labels to be []map[string]interface{}, got %T", set["labels"])
	}
	if len(labels) != 2 {
		t.Fatalf("expected 2 labels, got %d", len(labels))
	}

	filters, ok := set["filters"].([]interface{})
	if !ok {
		t.Fatalf("expected filters to be []interface{}, got %T", set["filters"])
	}
	if len(filters) != 0 {
		t.Fatalf("expected filters to be empty, got %d items", len(filters))
	}

	lookup := map[string][]string{}
	for _, label := range labels {
		field, _ := label["field"].(string)
		values, _ := label["values"].([]string)
		lookup[field] = values
	}

	if got := len(lookup["k8s.container.name"]); got != 1 || lookup["k8s.container.name"][0] != "credit-card-order-service" {
		t.Fatalf("unexpected values for k8s.container.name: %#v", lookup["k8s.container.name"])
	}

	if got := len(lookup["dt.kubernetes.workload.name"]); got != 1 || lookup["dt.kubernetes.workload.name"][0] != "credit-card-order-service" {
		t.Fatalf("unexpected values for dt.kubernetes.workload.name: %#v", lookup["dt.kubernetes.workload.name"])
	}
}
