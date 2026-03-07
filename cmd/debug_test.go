package cmd

import "testing"

func TestParseFilters(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		want    map[string][]string
	}{
		{
			name:  "single filter",
			input: "k8s.namespace.name=prod",
			want: map[string][]string{
				"k8s.namespace.name": {"prod"},
			},
		},
		{
			name:  "multiple filters and duplicate keys",
			input: "k8s.namespace.name=prod, dt.entity.host=HOST-1,k8s.namespace.name=stage",
			want: map[string][]string{
				"k8s.namespace.name": {"prod", "stage"},
				"dt.entity.host":     {"HOST-1"},
			},
		},
		{name: "missing equals", input: "k8s.namespace.name", wantErr: true},
		{name: "empty input", input: "", wantErr: true},
		{name: "empty key", input: "=prod", wantErr: true},
		{name: "empty value", input: "k8s.namespace.name=", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFilters(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(got) != len(tt.want) {
				t.Fatalf("unexpected map size: got=%d want=%d", len(got), len(tt.want))
			}

			for key, expectedValues := range tt.want {
				values, ok := got[key]
				if !ok {
					t.Fatalf("missing key %q", key)
				}
				if len(values) != len(expectedValues) {
					t.Fatalf("unexpected values length for %q: got=%d want=%d", key, len(values), len(expectedValues))
				}
				for i := range values {
					if values[i] != expectedValues[i] {
						t.Fatalf("unexpected value at %q[%d]: got=%q want=%q", key, i, values[i], expectedValues[i])
					}
				}
			}
		})
	}
}

func TestParseBreakpoint(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantFile string
		wantLine int
		wantErr  bool
	}{
		{name: "valid", input: "OrderController.java:306", wantFile: "OrderController.java", wantLine: 306},
		{name: "valid with spaces", input: " OrderController.java : 306 ", wantFile: "OrderController.java", wantLine: 306},
		{name: "missing separator", input: "OrderController.java", wantErr: true},
		{name: "empty file", input: ":123", wantErr: true},
		{name: "non numeric line", input: "OrderController.java:abc", wantErr: true},
		{name: "non positive line", input: "OrderController.java:0", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileName, lineNumber, err := parseBreakpoint(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if fileName != tt.wantFile {
				t.Fatalf("unexpected file name: got=%q want=%q", fileName, tt.wantFile)
			}

			if lineNumber != tt.wantLine {
				t.Fatalf("unexpected line number: got=%d want=%d", lineNumber, tt.wantLine)
			}
		})
	}
}
