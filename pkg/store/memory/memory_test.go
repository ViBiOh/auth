package memory

import (
	"errors"
	"flag"
	"reflect"
	"strings"
	"testing"

	"github.com/ViBiOh/auth/v2/pkg/model"
)

func TestFlags(t *testing.T) {
	var cases = []struct {
		intention string
		want      string
	}{
		{
			"simple",
			"Usage of simple:\n  -profiles string\n    \t[memory] Users profiles in the form 'id:profile1|profile2,id2:profile1' {SIMPLE_PROFILES}\n  -users string\n    \t[memory] Users credentials in the form 'id:login:password,id2:login2:password2' {SIMPLE_USERS}\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			fs := flag.NewFlagSet(testCase.intention, flag.ContinueOnError)
			Flags(fs, "")

			var writer strings.Builder
			fs.SetOutput(&writer)
			fs.Usage()

			if got := writer.String(); got != testCase.want {
				t.Errorf("Flags() = `%s`, want `%s`", got, testCase.want)
			}
		})
	}
}

func TestLoadIdent(t *testing.T) {
	type args struct {
		ident string
	}

	var cases = []struct {
		intention string
		args      args
		want      map[string]basicUser
		wantErr   error
	}{
		{
			"empty",
			args{
				ident: "",
			},
			nil,
			nil,
		},
		{
			"invalid format",
			args{
				ident: "1:vibioh",
			},
			nil,
			errors.New("invalid format for user ident `1:vibioh`"),
		},
		{
			"invalid number",
			args{
				ident: "first:vibioh:secret",
			},
			nil,
			errors.New("strconv.ParseUint: parsing \"first\": invalid syntax"),
		},
		{
			"same id",
			args{
				ident: "1:vibioh:secret,1:guest:password",
			},
			nil,
			errors.New("id already exists for user ident `1:guest:password`"),
		},
		{
			"multiple",
			args{
				ident: "1:VIBIOH:secret,2:guest:password",
			},
			map[string]basicUser{
				"vibioh": {
					model.NewUser(1, "vibioh"),
					[]byte("secret"),
				},
				"guest": {
					model.NewUser(2, "guest"),
					[]byte("password"),
				},
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			got, gotErr := loadIdent(tc.args.ident)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && gotErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("loadIdent() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}

func TestLoadAuth(t *testing.T) {
	type args struct {
		auth string
	}

	var cases = []struct {
		intention string
		args      args
		want      map[uint64][]string
		wantErr   error
	}{
		{
			"empty",
			args{
				auth: "",
			},
			nil,
			nil,
		},
		{
			"invalid format",
			args{
				auth: "admin",
			},
			nil,
			errors.New("invalid format of user auth `admin`"),
		},
		{
			"invalid number",
			args{
				auth: "first:admin",
			},
			nil,
			errors.New("strconv.ParseUint: parsing \"first\": invalid syntax"),
		},
		{
			"multiple",
			args{
				auth: "1:admin|user,2:guest",
			},
			map[uint64][]string{
				1: {"admin", "user"},
				2: {"guest"},
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			got, gotErr := loadAuth(tc.args.auth)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && gotErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("loadAuth() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}
