package main

import (
	"reflect"
	"testing"
)

func Test_parseURL(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *service
		wantErr bool
	}{
		{name: "test1",
			args: args{path: "http://1.1.1.1"},
			want: &service{
				Scheme: "http",
				Host:   "1.1.1.1",
				Port:   "",
			},
			wantErr: false,
		},
		{name: "test2",
			args: args{path: "http://1.1.1.1:80"},
			want: &service{
				Scheme: "http",
				Host:   "1.1.1.1",
				Port:   "80",
			},
			wantErr: false,
		},
		{name: "test3",
			args: args{path: "http://1.1.1.1:234234"},
			want: &service{
				Scheme: "http",
				Host:   "1.1.1.1",
				Port:   "234234",
			},
			wantErr: false,
		},
		{name: "test4",
			args: args{path: "http://www.demo.com"},
			want: &service{
				Scheme: "http",
				Host:   "www.demo.com",
				Port:   "",
			},
			wantErr: false,
		},
		{name: "test5",
			args: args{path: "https://www.demo.com"},
			want: &service{
				Scheme: "https",
				Host:   "www.demo.com",
				Port:   "",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseURL(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseURL() got = %v, want %v", got, tt.want)
			}
		})
	}
}
