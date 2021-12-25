package querysan

import "testing"

func Test_timezone(t *testing.T) {
	t.Skip()
	tests := []struct {
		name  string
		want  string
		want1 int
	}{
		{
			"Simple",
			"JST",
			9 * 60 * 60,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := timezone()
			if got != tt.want {
				t.Errorf("timezone() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("timezone() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
