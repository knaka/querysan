package querysan

import (
	"reflect"
	"testing"
)

func Test_tokenize(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			"Simple",
			args{
				"寿司が食べたい。\n佐藤さんと、もっとHello World したい。",
			},
			[]string{
				"寿司", "が", "食べ", "たい",
				"佐藤", "さん", "と", "もっと", "hello", "world", "し", "たい",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := words(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("words() = %v, want %v", got, tt.want)
			}
		})
	}
}
