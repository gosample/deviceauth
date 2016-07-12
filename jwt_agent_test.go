// Copyright 2016 Mender Software AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.
package main

import (
	"crypto/rsa"
	"github.com/mendersoftware/deviceauth/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewJWTAgent(t *testing.T) {
	testCases := []struct {
		privKeyPath string
		privKey     *rsa.PrivateKey
		issuer      string
		err         string
	}{
		{
			privKeyPath: "testdata/private.pem",
			privKey:     test.LoadPrivKey("testdata/private.pem", t),
			issuer:      "Mender",
			err:         "",
		},
		{
			privKeyPath: "wrong_path",
			privKey:     nil,
			issuer:      "",
			err:         ErrMsgPrivKeyReadFailed + ": open wrong_path: no such file or directory",
		},
		{
			privKeyPath: "testdata/private_broken.pem",
			privKey:     nil,
			issuer:      "",
			err:         ErrMsgPrivKeyNotPEMEncoded,
		},
		{
			privKeyPath: "testdata/public.pem",
			privKey:     nil,
			issuer:      "",
			err:         "unknown server private key type; got: PUBLIC KEY, want: RSA PRIVATE KEY",
		},
	}

	for _, tc := range testCases {
		c := JWTAgentConfig{
			ServerPrivKeyPath: tc.privKeyPath,
			Issuer:            tc.issuer,
		}

		jwt, err := NewJWTAgent(c)
		if tc.err != "" {
			assert.EqualError(t, err, tc.err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tc.privKey, jwt.privKey)
			assert.Equal(t, tc.issuer, jwt.issuer)
		}
	}
}

func TestGenerateTokenSignRS256(t *testing.T) {
	testCases := []struct {
		privKeyPath string
		devId       string
	}{
		{
			privKeyPath: "testdata/private.pem",
			devId:       "deviceId",
		},
	}
	for _, tc := range testCases {
		c := JWTAgentConfig{
			ServerPrivKeyPath: tc.privKeyPath,
			Issuer:            "Mender",
			ExpirationTimeout: 1,
		}
		jwt, err := NewJWTAgent(c)
		assert.NoError(t, err)
		token, err := jwt.GenerateTokenSignRS256(tc.devId)
		assert.NoError(t, err)
		assert.Equal(t, tc.devId, token.DevId)
	}
}

func TestValidateTokenSignRS256(t *testing.T) {
	testCases := []struct {
		privKeyPath string
		devId       string
		expiration  int64
		err         error
		delay       int64
	}{
		{
			privKeyPath: "testdata/private.pem",
			devId:       "deviceId",
			expiration:  time.Now().Unix() + 123,
			err:         nil,
		},
		{
			privKeyPath: "testdata/private.pem",
			devId:       "deviceId",
			expiration:  0,
			err:         ErrTokenExpired,
			delay:       1,
		},
	}
	for _, tc := range testCases {
		c := JWTAgentConfig{
			ServerPrivKeyPath: tc.privKeyPath,
			Issuer:            "Mender",
			ExpirationTimeout: tc.expiration,
		}
		jwt, err := NewJWTAgent(c)
		assert.NoError(t, err)
		token, err := jwt.GenerateTokenSignRS256(tc.devId)
		assert.NoError(t, err)
		assert.Equal(t, tc.devId, token.DevId)
		if tc.err == ErrTokenExpired {
			time.Sleep(time.Second * time.Duration(tc.delay))
		}
		_, err = jwt.ValidateTokenSignRS256(token.Token)
		if tc.err != nil {
			assert.EqualError(t, err, tc.err.Error())
		} else {
			assert.NoError(t, err)
		}
		// assert jit is uuid v4?
	}
}