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
		want    Url
		wantErr bool
	}{
		{name: "test1",
			args: args{path: "http://1.1.1.1"},
			want: Url{
				Scheme: "http",
				Host:   "1.1.1.1",
				Port:   "80",
			},
			wantErr: false,
		},
		{name: "test2",
			args: args{path: "http://1.1.1.1:80"},
			want: Url{
				Scheme: "http",
				Host:   "1.1.1.1",
				Port:   "80",
			},
			wantErr: false,
		},
		{name: "test3",
			args: args{path: "http://1.1.1.1:234234"},
			want: Url{
				Scheme: "http",
				Host:   "1.1.1.1",
				Port:   "234234",
			},
			wantErr: false,
		},
		{name: "test4",
			args: args{path: "http://www.demo.com"},
			want: Url{
				Scheme: "http",
				Host:   "www.demo.com",
				Port:   "80",
			},
			wantErr: false,
		},
		{name: "test5",
			args: args{path: "https://www.demo.com"},
			want: Url{
				Scheme: "https",
				Host:   "www.demo.com",
				Port:   "443",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePathToUrl(tt.args.path)
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
