package stringutil

import (
	"reflect"
	"testing"
)

func TestChunksOfN(t *testing.T) {
	type args struct {
		s         string
		chunkSize int
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "StringIsMultipleOf5_Success",
			args: args{
				s:         "hello world how are you doing today I'm doing super great!!!",
				chunkSize: 5,
			},
			want: []string{"hello", " worl", "d how", " are ", "you d", "oing ", "today", " I'm ", "doing", " supe", "r gre", "at!!!"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := chunksOfN(tt.args.s, tt.args.chunkSize); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("chunksOfN() = %v, want %v", got, tt.want)
			}
		})
	}
}
