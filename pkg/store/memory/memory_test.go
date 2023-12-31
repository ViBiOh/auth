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
	t.Parallel()

	cases := map[string]struct {
		want string
	}{
		"simple": {
			"Usage of simple:\n  -profiles string slice\n    \t[memory] Users profiles in the form 'id:profile1|profile2' ${SIMPLE_PROFILES}, as a string slice, environment variable separated by \",\"\n  -users string slice\n    \t[memory] Users credentials in the form 'id:login:password' ${SIMPLE_USERS}, as a string slice, environment variable separated by \"|\"\n",
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			fs := flag.NewFlagSet(intention, flag.ContinueOnError)
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
	t.Parallel()

	type args struct {
		idents []string
	}

	cases := map[string]struct {
		args    args
		want    map[string]basicUser
		wantErr error
	}{
		"empty": {
			args{},
			nil,
			nil,
		},
		"invalid format": {
			args{
				idents: []string{"1:vibioh"},
			},
			nil,
			errors.New("invalid format for user ident `1:vibioh`"),
		},
		"invalid number": {
			args{
				idents: []string{"first:vibioh:secret"},
			},
			nil,
			errors.New("strconv.ParseUint: parsing \"first\": invalid syntax"),
		},
		"same id": {
			args{
				idents: []string{"1:vibioh:secret", "1:guest:password"},
			},
			nil,
			errors.New("id already exists for user ident `1:guest:password`"),
		},
		"multiple": {
			args{
				idents: []string{"1:VIBIOH:secret", "2:guest:password"},
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

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			got, gotErr := loadIdent(testCase.args.idents)

			failed := false

			if testCase.wantErr == nil && gotErr != nil {
				failed = true
			} else if testCase.wantErr != nil && gotErr == nil {
				failed = true
			} else if testCase.wantErr != nil && gotErr != nil && !strings.Contains(gotErr.Error(), testCase.wantErr.Error()) {
				failed = true
			} else if !reflect.DeepEqual(got, testCase.want) {
				failed = true
			}

			if failed {
				t.Errorf("loadIdent() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestLoadAuth(t *testing.T) {
	t.Parallel()

	type args struct {
		auths []string
	}

	cases := map[string]struct {
		args    args
		want    map[uint64][]string
		wantErr error
	}{
		"empty": {
			args{},
			nil,
			nil,
		},
		"invalid format": {
			args{
				auths: []string{"admin"},
			},
			nil,
			errors.New("invalid format of user auth `admin`"),
		},
		"invalid number": {
			args{
				auths: []string{"first:admin"},
			},
			nil,
			errors.New("strconv.ParseUint: parsing \"first\": invalid syntax"),
		},
		"multiple": {
			args{
				auths: []string{"1:admin|user", "2:guest"},
			},
			map[uint64][]string{
				1: {"admin", "user"},
				2: {"guest"},
			},
			nil,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			got, gotErr := loadAuth(testCase.args.auths)

			failed := false

			if testCase.wantErr == nil && gotErr != nil {
				failed = true
			} else if testCase.wantErr != nil && gotErr == nil {
				failed = true
			} else if testCase.wantErr != nil && gotErr != nil && !strings.Contains(gotErr.Error(), testCase.wantErr.Error()) {
				failed = true
			} else if !reflect.DeepEqual(got, testCase.want) {
				failed = true
			}

			if failed {
				t.Errorf("loadAuth() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, testCase.want, testCase.wantErr)
			}
		})
	}
}
