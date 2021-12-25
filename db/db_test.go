package db

import (
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestAddIndex(t *testing.T) {
	dbPath := "/tmp/test.db"
	if err := os.Remove(dbPath); err != nil {
		t.Fail()
	}
	if err := Initialize(dbPath); err != nil {
		t.Fail()
	}
	if err := UpsertEntry("data/1.txt"); err != nil {
		t.Fail()
	}
	if err := UpsertEntry("data/2.txt"); err != nil {
		t.Fail()
	}
	paths, err := Query("適*")
	if err != nil {
		t.Fail()
	}
	fmt.Println(paths)
	if err := DeleteEntry("data/1.txt"); err != nil {
		t.Fail()
	}
	Finalize()
}

func Test_fromExternalDateTime(t *testing.T) {
	type args struct {
		datetime string
	}
	datetime := "2021-12-24T13:36:35.891653562Z"
	time1 := time.Date(2021, 12, 24, 13, 36, 35, 891653562, time.UTC).Local()
	tests := []struct {
		name    string
		args    args
		want    *time.Time
		wantErr bool
	}{
		{
			"Basic",
			args{
				datetime,
			},
			&time1,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fromExternalTime(tt.args.datetime)
			if (err != nil) != tt.wantErr {
				t.Errorf("fromExternalTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fromExternalTime() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_toExternalDateTime(t *testing.T) {
	type args struct {
		time1 *time.Time
	}
	datetime := "2021-12-24T13:36:35.891653562Z"
	time1 := time.Date(2021, 12, 24, 13, 36, 35, 891653562, time.UTC).Local()
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"Simple",
			args{&time1},
			datetime,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toExternalTime(tt.args.time1); got != tt.want {
				t.Errorf("toExternalTime() = %v, want %v", got, tt.want)
			}
		})
	}
}
