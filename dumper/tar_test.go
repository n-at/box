package dumper

import "testing"

func Test_splitTargetPath(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name       string
		args       args
		directory  string
		targetFile string
	}{
		{
			name: "normal",
			args: args{
				path: "/directory1/directory2/file",
			},
			directory:  "/directory1/directory2",
			targetFile: "file",
		}, {
			name: "ends with slash",
			args: args{
				path: "/d1/d2/",
			},
			directory:  "/d1",
			targetFile: "d2",
		}, {
			name: "double slash",
			args: args{
				path: "/d1//d2/d3",
			},
			directory:  "/d1//d2",
			targetFile: "d3",
		}, {
			name: "at root",
			args: args{
				path: "/root",
			},
			directory:  "/",
			targetFile: "root",
		}, {
			name: "ends with double slash",
			args: args{
				path: "/d1/d2//",
			},
			directory:  "/d1",
			targetFile: "d2",
		}, {
			name: "relative path",
			args: args{
				path: "d1/d2/d3",
			},
			directory:  "d1/d2",
			targetFile: "d3",
		}, {
			name: "relative single",
			args: args{
				path: "dir",
			},
			directory:  "",
			targetFile: "dir",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := splitTargetPath(tt.args.path)
			if got != tt.directory {
				t.Errorf("splitTargetPath() directory = %v, want %v", got, tt.directory)
			}
			if got1 != tt.targetFile {
				t.Errorf("splitTargetPath() targetFile = %v, want %v", got1, tt.targetFile)
			}
		})
	}
}
