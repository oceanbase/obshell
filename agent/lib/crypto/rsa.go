/*
 * Copyright (c) 2024 OceanBase.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"

	"github.com/oceanbase/obshell/agent/errors"
)

func sectionalEncrypt(raw []byte, pub *rsa.PublicKey) (string, error) {
	// Sectional encryption.
	blockSize := KEY_SIZE/8 - 11
	numBlocks := (len(raw) + blockSize - 1) / blockSize
	ciphertext := make([]byte, 0)
	for i := 0; i < numBlocks; i++ {
		start := i * blockSize
		end := (i + 1) * blockSize
		if end > len(raw) {
			end = len(raw)
		}
		encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, pub, raw[start:end])
		if err != nil {
			return "", err
		}
		ciphertext = append(ciphertext, encrypted...)
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func RSAEncrypt(raw []byte, pk string) (string, error) {
	pkix, err := base64.StdEncoding.DecodeString(pk)
	if err != nil {
		return "", err
	}
	pub, err := x509.ParsePKCS1PublicKey(pkix)
	if err != nil {
		return "", err
	}
	if len(raw) == 0 {
		b, err := rsa.EncryptPKCS1v15(rand.Reader, pub, raw)
		return base64.StdEncoding.EncodeToString(b), err
	}

	return sectionalEncrypt(raw, pub)
}

type RSACrypto struct {
	pk   *rsa.PrivateKey
	spri string
	spub string
}

const (
	KEY_SIZE    = 512
	ERR_NO_INIT = "rsa crypto not initialized"
)

func NewRSACrypto() (*RSACrypto, error) {
	pk, err := rsa.GenerateKey(rand.Reader, KEY_SIZE)
	if err != nil {
		return nil, err
	}
	mpri := x509.MarshalPKCS1PrivateKey(pk)
	mpub := x509.MarshalPKCS1PublicKey(&pk.PublicKey)
	return &RSACrypto{
		pk:   pk,
		spri: base64.StdEncoding.EncodeToString(mpri),
		spub: base64.StdEncoding.EncodeToString(mpub),
	}, nil
}

func NewRSACryptoFromKey(privateKey string) (*RSACrypto, error) {
	// Parameter `privateKey` is assumed to be base64-encoded
	mpri, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		return nil, err
	}
	pk, err := x509.ParsePKCS1PrivateKey(mpri)
	if err != nil {
		return nil, err
	}
	mpub := x509.MarshalPKCS1PublicKey(&pk.PublicKey)
	return &RSACrypto{
		pk:   pk,
		spri: privateKey,
		spub: base64.StdEncoding.EncodeToString(mpub),
	}, nil
}

func (r *RSACrypto) Encrypt(raw string) (string, error) {
	if r == nil || r.pk == nil {
		return "", errors.Occur(errors.ErrCommonUnexpected, ERR_NO_INIT)
	}
	return sectionalEncrypt([]byte(raw), &r.pk.PublicKey)
}

func (r *RSACrypto) DecryptAndReturnBytes(raw string) ([]byte, error) {
	if r == nil || r.pk == nil {
		return nil, errors.Occur(errors.ErrCommonUnexpected, ERR_NO_INIT)
	}
	decRaw, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, err
	}

	// Sectional decryption.
	ciphertext := make([]byte, 0)
	keySize := r.pk.Size()
	numBlocks := (len(decRaw) + keySize - 1) / keySize
	for i := 0; i < numBlocks; i++ {
		start := i * keySize
		end := (i + 1) * keySize
		if end > len(decRaw) {
			end = len(decRaw)
		}
		decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, r.pk, decRaw[start:end])
		if err != nil {
			return nil, err
		}
		ciphertext = append(ciphertext, decrypted...)
	}
	return ciphertext, err
}

func (r *RSACrypto) Decrypt(raw string) (string, error) {
	ciphertext, err := r.DecryptAndReturnBytes(raw)
	if err != nil {
		return "", err
	}
	return string(ciphertext), err
}

func (r *RSACrypto) TryEncrypt(raw string) string {
	s, err := r.Encrypt(raw)
	if err != nil {
		return raw
	}
	return s
}

func (r *RSACrypto) TryDecrypt(raw string) string {
	s, err := r.Decrypt(raw)
	if err != nil {
		return raw
	}
	return s
}

func (r *RSACrypto) Private() string {
	if r == nil || r.pk == nil {
		return ""
	}
	return r.spri
}

func (r *RSACrypto) Public() string {
	if r == nil || r.pk == nil {
		return ""
	}
	return r.spub
}
