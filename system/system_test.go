package system

import (
	"reflect"
	"testing"
)

func TestGetAllIP(t *testing.T) {
	tests := []struct {
		name    string
		want    map[string][]string
		wantErr bool
	}{
		{
			"获取所有IP正向测试用例",
			map[string][]string{"IPV4": {"10.1.2.153"}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetAllIP()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// TODO: 这个断言有问题，实际获取到各种IP比测试用例要多的
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAllIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVerifyPortIsUnused(t *testing.T) {
	type args struct {
		port uint16
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			"端口未被占用场景",
			args{port: 19090},
			true,
			false,
		},
		{
			"端口被占用场景",
			args{port: 9528},
			false,
			false,
		},
		{
			"端口被占用场景，但无权限使用",
			args{port: 80},
			false,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := VerifyPortIsUnused(tt.args.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyPortIsUnused() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("VerifyPortIsUnused() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVerifyRuntimeEnv(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		{
			"验证Host运行环境",
			"host",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := VerifyRuntimeEnv()
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyRuntimeEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("VerifyRuntimeEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}
