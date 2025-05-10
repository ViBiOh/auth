package argon_test

import (
	"errors"
	"testing"

	"github.com/ViBiOh/auth/v2/pkg/argon"
	"github.com/stretchr/testify/assert"
)

func TestReversible(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		password := "correct horse battery staple"

		encodedHash, err := argon.GenerateFromPassword(password)
		assert.NoError(t, err)

		err = argon.CompareHashAndPassword(encodedHash, password)
		assert.NoError(t, err)
	})
}

func TestCompareHashAndPassowrd(t *testing.T) {
	t.Parallel()

	type args struct {
		content string
	}

	cases := map[string]struct {
		args    args
		wantErr error
	}{
		"empty": {
			args{
				content: "",
			},
			argon.ErrInvalidEncodedHash,
		},
		"not argon": {
			args{
				content: "$bcrypt$v=19$m=7168,t=5,p=1$OOcpl5JoznYGLWXiCvejrHnj4KaHp0I$efHnbPwEcvonAQrQR8xBq6X7GgIEHRuii0DRM0egXZM",
			},
			argon.ErrUnhandledEncodedHash,
		},
		"parse version": {
			args{
				content: "$argon2id$v=a$m=7168,t=5,p=1$OOcpl5JoznYGLWXiCvejrHnj4KaHp0I$efHnbPwEcvonAQrQR8xBq6X7GgIEHRuii0DRM0egXZM",
			},
			errors.New("parse: decode version"),
		},
		"invalid version": {
			args{
				content: "$argon2id$v=1$m=7168,t=5,p=1$OOcpl5JoznYGLWXiCvejrHnj4KaHp0I$efHnbPwEcvonAQrQR8xBq6X7GgIEHRuii0DRM0egXZM",
			},
			argon.ErrUnhandledVersion,
		},
		"parse param": {
			args{
				content: "$argon2id$v=19$m=a,t=b,p=c$OOcpl5JoznYGLWXiCvejrHnj4KaHp0I$efHnbPwEcvonAQrQR8xBq6X7GgIEHRuii0DRM0egXZM",
			},
			errors.New("parse: decode params"),
		},
		"decode salt": {
			args{
				content: "$argon2id$v=19$m=7168,t=5,p=1$OOcpl5JoznYGLWXiCvejrHnj4KaHp0I=$efHnbPwEcvonAQrQR8xBq6X7GgIEHRuii0DRM0egXZM",
			},
			errors.New("parse: decode salt"),
		},
		"decode hash": {
			args{
				content: "$argon2id$v=19$m=7168,t=5,p=1$OOcpl5JoznYGLWXiCvejrHnj4KaHp0I$efHnbPwEcvonAQrQR8xBq6X7GgIEHRuii0DRM0egXZM=",
			},
			errors.New("parse: decode hash"),
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			gotErr := argon.CompareHashAndPassword(testCase.args.content, "")

			assert.ErrorContains(t, gotErr, testCase.wantErr.Error())
		})
	}
}
