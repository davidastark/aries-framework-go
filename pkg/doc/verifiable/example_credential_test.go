/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package verifiable_test

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/ed25519signature2018"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
)

type UniversityDegree struct {
	Type       string `json:"type,omitempty"`
	University string `json:"university,omitempty"`
}

type UniversityDegreeSubject struct {
	ID     string           `json:"id,omitempty"`
	Name   string           `json:"name,omitempty"`
	Spouse string           `json:"spouse,omitempty"`
	Degree UniversityDegree `json:"degree,omitempty"`
}

type UniversityDegreeCredential struct {
	*verifiable.Credential

	ReferenceNumber int `json:"referenceNumber,omitempty"`
}

func (udc *UniversityDegreeCredential) MarshalJSON() ([]byte, error) {
	// todo too complex! (https://github.com/hyperledger/aries-framework-go/issues/847)
	c := udc.Credential
	cp := *c

	cp.CustomFields = map[string]interface{}{
		"referenceNumber": udc.ReferenceNumber,
	}

	return json.Marshal(&cp)
}

//nolint:gochecknoglobals
var (
	// Private key generated by ed25519.GenerateKey(rand.Reader)
	issuerPrivKey = ed25519.PrivateKey{72, 67, 163, 188, 235, 199, 239, 146, 129, 52, 228, 34, 44, 106, 23, 144, 189, 57, 115, 171, 4, 217, 54, 121, 41, 155, 251, 83, 1, 240, 238, 65, 234, 100, 192, 93, 251, 181, 198, 73, 122, 220, 27, 48, 93, 73, 166, 33, 152, 140, 168, 36, 9, 205, 59, 161, 137, 7, 164, 9, 176, 252, 1, 171} //nolint:lll
	issuerPubKey  = ed25519.PublicKey{234, 100, 192, 93, 251, 181, 198, 73, 122, 220, 27, 48, 93, 73, 166, 33, 152, 140, 168, 36, 9, 205, 59, 161, 137, 7, 164, 9, 176, 252, 1, 171}                                                                                                                                                   //nolint:lll
	issued        = time.Date(2010, time.January, 1, 19, 23, 24, 0, time.UTC)
	expired       = time.Date(2020, time.January, 1, 19, 23, 24, 0, time.UTC)
)

//nolint:lll
func ExampleCredential_embedding() {
	vc := &UniversityDegreeCredential{
		Credential: &verifiable.Credential{
			Context: []string{
				"https://www.w3.org/2018/credentials/v1",
				"https://www.w3.org/2018/credentials/examples/v1"},
			ID: "http://example.edu/credentials/1872",
			Types: []string{
				"VerifiableCredential",
				"UniversityDegreeCredential"},
			Subject: UniversityDegreeSubject{
				ID:     "did:example:ebfeb1f712ebc6f1c276e12ec21",
				Name:   "Jayden Doe",
				Spouse: "did:example:c276e12ec21ebfeb1f712ebc6f1",
				Degree: UniversityDegree{
					Type:       "BachelorDegree",
					University: "MIT",
				},
			},
			Issuer: verifiable.Issuer{
				ID:   "did:example:76e12ec712ebc6f1c221ebfeb1f",
				Name: "Example University",
			},
			Issued:  &issued,
			Expired: &expired,
			Schemas: []verifiable.TypedID{},
		},
		ReferenceNumber: 83294847,
	}

	// Marshal to JSON to verify the result of decoding.
	vcBytes, err := json.Marshal(vc)
	if err != nil {
		fmt.Println("failed to marshal VC to JSON")
	}

	fmt.Println(string(vcBytes))

	// Marshal to JWS.
	jwtClaims, err := vc.JWTClaims(true)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to marshal JWT claims of VC: %w", err))
	}

	jws, err := jwtClaims.MarshalJWS(verifiable.EdDSA, getSigner(issuerPrivKey), "")
	if err != nil {
		fmt.Println(fmt.Errorf("failed to sign VC inside JWT: %w", err))
	}

	fmt.Println(jws)

	// Decode JWS and make sure it's coincide with JSON.
	_, vcBytesFromJWS, err := verifiable.NewCredential(
		[]byte(jws),
		verifiable.WithPublicKeyFetcher(verifiable.SingleKey(issuerPubKey, kms.Ed25519Type)))
	if err != nil {
		fmt.Println(fmt.Errorf("failed to encode VC from JWS: %w", err))
	}

	// todo missing referenceNumber here (https://github.com/hyperledger/aries-framework-go/issues/847)
	fmt.Println(string(vcBytesFromJWS))

	// Output:
	// {"@context":["https://www.w3.org/2018/credentials/v1","https://www.w3.org/2018/credentials/examples/v1"],"credentialSchema":[],"credentialSubject":{"degree":{"type":"BachelorDegree","university":"MIT"},"id":"did:example:ebfeb1f712ebc6f1c276e12ec21","name":"Jayden Doe","spouse":"did:example:c276e12ec21ebfeb1f712ebc6f1"},"expirationDate":"2020-01-01T19:23:24Z","id":"http://example.edu/credentials/1872","issuanceDate":"2010-01-01T19:23:24Z","issuer":{"id":"did:example:76e12ec712ebc6f1c221ebfeb1f","name":"Example University"},"referenceNumber":83294847,"type":["VerifiableCredential","UniversityDegreeCredential"]}
	// eyJhbGciOiJFZERTQSIsImtpZCI6IiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1Nzc5MDY2MDQsImlhdCI6MTI2MjM3MzgwNCwiaXNzIjoiZGlkOmV4YW1wbGU6NzZlMTJlYzcxMmViYzZmMWMyMjFlYmZlYjFmIiwianRpIjoiaHR0cDovL2V4YW1wbGUuZWR1L2NyZWRlbnRpYWxzLzE4NzIiLCJuYmYiOjEyNjIzNzM4MDQsInN1YiI6ImRpZDpleGFtcGxlOmViZmViMWY3MTJlYmM2ZjFjMjc2ZTEyZWMyMSIsInZjIjp7IkBjb250ZXh0IjpbImh0dHBzOi8vd3d3LnczLm9yZy8yMDE4L2NyZWRlbnRpYWxzL3YxIiwiaHR0cHM6Ly93d3cudzMub3JnLzIwMTgvY3JlZGVudGlhbHMvZXhhbXBsZXMvdjEiXSwiY3JlZGVudGlhbFNjaGVtYSI6W10sImNyZWRlbnRpYWxTdWJqZWN0Ijp7ImRlZ3JlZSI6eyJ0eXBlIjoiQmFjaGVsb3JEZWdyZWUiLCJ1bml2ZXJzaXR5IjoiTUlUIn0sImlkIjoiZGlkOmV4YW1wbGU6ZWJmZWIxZjcxMmViYzZmMWMyNzZlMTJlYzIxIiwibmFtZSI6IkpheWRlbiBEb2UiLCJzcG91c2UiOiJkaWQ6ZXhhbXBsZTpjMjc2ZTEyZWMyMWViZmViMWY3MTJlYmM2ZjEifSwiaXNzdWVyIjp7Im5hbWUiOiJFeGFtcGxlIFVuaXZlcnNpdHkifSwidHlwZSI6WyJWZXJpZmlhYmxlQ3JlZGVudGlhbCIsIlVuaXZlcnNpdHlEZWdyZWVDcmVkZW50aWFsIl19fQ.KhUlkaZwjIiaMzos9FfxLeuIWmJJIPcb_uKz8amzP79ZaKeKwjclYT82q2QWUUObaT2-eB42H5kjycdSKg3tCw
	// {"@context":["https://www.w3.org/2018/credentials/v1","https://www.w3.org/2018/credentials/examples/v1"],"credentialSchema":[],"credentialSubject":{"degree":{"type":"BachelorDegree","university":"MIT"},"id":"did:example:ebfeb1f712ebc6f1c276e12ec21","name":"Jayden Doe","spouse":"did:example:c276e12ec21ebfeb1f712ebc6f1"},"expirationDate":"2020-01-01T19:23:24Z","id":"http://example.edu/credentials/1872","issuanceDate":"2010-01-01T19:23:24Z","issuer":{"id":"did:example:76e12ec712ebc6f1c221ebfeb1f","name":"Example University"},"type":["VerifiableCredential","UniversityDegreeCredential"]}
}

//nolint:lll
func ExampleCredential_extraFields() {
	vc := &verifiable.Credential{
		Context: []string{
			"https://www.w3.org/2018/credentials/v1",
			"https://www.w3.org/2018/credentials/examples/v1"},
		ID: "http://example.edu/credentials/1872",
		Types: []string{
			"VerifiableCredential",
			"UniversityDegreeCredential"},
		Subject: UniversityDegreeSubject{
			ID:     "did:example:ebfeb1f712ebc6f1c276e12ec21",
			Name:   "Jayden Doe",
			Spouse: "did:example:c276e12ec21ebfeb1f712ebc6f1",
			Degree: UniversityDegree{
				Type:       "BachelorDegree",
				University: "MIT",
			},
		},
		Issuer: verifiable.Issuer{
			ID:   "did:example:76e12ec712ebc6f1c221ebfeb1f",
			Name: "Example University",
		},
		Issued:  &issued,
		Expired: &expired,
		Schemas: []verifiable.TypedID{},
		CustomFields: map[string]interface{}{
			"referenceNumber": 83294847,
		},
	}

	// Marshal to JSON.
	vcBytes, err := json.Marshal(vc)
	if err != nil {
		fmt.Println("failed to marshal VC to JSON")
	}

	fmt.Println(string(vcBytes))

	// Marshal to JWS.
	jwtClaims, err := vc.JWTClaims(true)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to marshal JWT claims of VC: %w", err))
	}

	jws, err := jwtClaims.MarshalJWS(verifiable.EdDSA, getSigner(issuerPrivKey), "")
	if err != nil {
		fmt.Println(fmt.Errorf("failed to sign VC inside JWT: %w", err))
	}

	fmt.Println(jws)

	// Decode JWS and make sure it's coincide with JSON.
	_, vcBytesFromJWS, err := verifiable.NewCredential(
		[]byte(jws),
		verifiable.WithPublicKeyFetcher(verifiable.SingleKey(issuerPubKey, kms.Ed25519Type)))
	if err != nil {
		fmt.Println(fmt.Errorf("failed to encode VC from JWS: %w", err))
	}

	fmt.Println(string(vcBytesFromJWS))

	// Output:
	// {"@context":["https://www.w3.org/2018/credentials/v1","https://www.w3.org/2018/credentials/examples/v1"],"credentialSchema":[],"credentialSubject":{"degree":{"type":"BachelorDegree","university":"MIT"},"id":"did:example:ebfeb1f712ebc6f1c276e12ec21","name":"Jayden Doe","spouse":"did:example:c276e12ec21ebfeb1f712ebc6f1"},"expirationDate":"2020-01-01T19:23:24Z","id":"http://example.edu/credentials/1872","issuanceDate":"2010-01-01T19:23:24Z","issuer":{"id":"did:example:76e12ec712ebc6f1c221ebfeb1f","name":"Example University"},"referenceNumber":83294847,"type":["VerifiableCredential","UniversityDegreeCredential"]}
	// eyJhbGciOiJFZERTQSIsImtpZCI6IiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1Nzc5MDY2MDQsImlhdCI6MTI2MjM3MzgwNCwiaXNzIjoiZGlkOmV4YW1wbGU6NzZlMTJlYzcxMmViYzZmMWMyMjFlYmZlYjFmIiwianRpIjoiaHR0cDovL2V4YW1wbGUuZWR1L2NyZWRlbnRpYWxzLzE4NzIiLCJuYmYiOjEyNjIzNzM4MDQsInN1YiI6ImRpZDpleGFtcGxlOmViZmViMWY3MTJlYmM2ZjFjMjc2ZTEyZWMyMSIsInZjIjp7IkBjb250ZXh0IjpbImh0dHBzOi8vd3d3LnczLm9yZy8yMDE4L2NyZWRlbnRpYWxzL3YxIiwiaHR0cHM6Ly93d3cudzMub3JnLzIwMTgvY3JlZGVudGlhbHMvZXhhbXBsZXMvdjEiXSwiY3JlZGVudGlhbFNjaGVtYSI6W10sImNyZWRlbnRpYWxTdWJqZWN0Ijp7ImRlZ3JlZSI6eyJ0eXBlIjoiQmFjaGVsb3JEZWdyZWUiLCJ1bml2ZXJzaXR5IjoiTUlUIn0sImlkIjoiZGlkOmV4YW1wbGU6ZWJmZWIxZjcxMmViYzZmMWMyNzZlMTJlYzIxIiwibmFtZSI6IkpheWRlbiBEb2UiLCJzcG91c2UiOiJkaWQ6ZXhhbXBsZTpjMjc2ZTEyZWMyMWViZmViMWY3MTJlYmM2ZjEifSwiaXNzdWVyIjp7Im5hbWUiOiJFeGFtcGxlIFVuaXZlcnNpdHkifSwicmVmZXJlbmNlTnVtYmVyIjo4LjMyOTQ4NDdlKzA3LCJ0eXBlIjpbIlZlcmlmaWFibGVDcmVkZW50aWFsIiwiVW5pdmVyc2l0eURlZ3JlZUNyZWRlbnRpYWwiXX19.IbGTRreDq5wZj3dETz6pnkRHdZOPwm9X-yQ-zyEve7bhOrZucqXRjmCLMkVFjNfcjJy5Xu4f3bcQe71pM--QCA
	// {"@context":["https://www.w3.org/2018/credentials/v1","https://www.w3.org/2018/credentials/examples/v1"],"credentialSchema":[],"credentialSubject":{"degree":{"type":"BachelorDegree","university":"MIT"},"id":"did:example:ebfeb1f712ebc6f1c276e12ec21","name":"Jayden Doe","spouse":"did:example:c276e12ec21ebfeb1f712ebc6f1"},"expirationDate":"2020-01-01T19:23:24Z","id":"http://example.edu/credentials/1872","issuanceDate":"2010-01-01T19:23:24Z","issuer":{"id":"did:example:76e12ec712ebc6f1c221ebfeb1f","name":"Example University"},"referenceNumber":83294847,"type":["VerifiableCredential","UniversityDegreeCredential"]}
}

//nolint:lll
func ExampleNewCredential() {
	// Issuer is about to issue the university degree credential for the Holder
	vcEncoded := &verifiable.Credential{
		Context: []string{
			"https://www.w3.org/2018/credentials/v1",
			"https://www.w3.org/2018/credentials/examples/v1"},
		ID: "http://example.edu/credentials/1872",
		Types: []string{
			"VerifiableCredential",
			"UniversityDegreeCredential"},
		Subject: UniversityDegreeSubject{
			ID:     "did:example:ebfeb1f712ebc6f1c276e12ec21",
			Name:   "Jayden Doe",
			Spouse: "did:example:c276e12ec21ebfeb1f712ebc6f1",
			Degree: UniversityDegree{
				Type:       "BachelorDegree",
				University: "MIT",
			},
		},
		Issuer: verifiable.Issuer{
			ID:   "did:example:76e12ec712ebc6f1c221ebfeb1f",
			Name: "Example University",
		},
		Issued:  &issued,
		Expired: &expired,
		Schemas: []verifiable.TypedID{},
		CustomFields: map[string]interface{}{
			"referenceNumber": 83294847,
		},
	}

	// ... in JWS form.
	jwtClaims, err := vcEncoded.JWTClaims(true)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to marshal JWT claims of VC: %w", err))
	}

	jws, err := jwtClaims.MarshalJWS(verifiable.EdDSA, getSigner(issuerPrivKey), "")
	if err != nil {
		fmt.Println(fmt.Errorf("failed to sign VC inside JWT: %w", err))
	}

	// The Holder receives JWS and decodes it.
	_, vcDecodedBytes, err := verifiable.NewCredential(
		[]byte(jws),
		verifiable.WithPublicKeyFetcher(verifiable.SingleKey(issuerPubKey, kms.Ed25519Type)))
	if err != nil {
		fmt.Println(fmt.Errorf("failed to decode VC JWS: %w", err))
	}

	// The Holder then e.g. can save the credential to her personal verifiable credential wallet.
	fmt.Println(string(vcDecodedBytes))

	// Output: {"@context":["https://www.w3.org/2018/credentials/v1","https://www.w3.org/2018/credentials/examples/v1"],"credentialSchema":[],"credentialSubject":{"degree":{"type":"BachelorDegree","university":"MIT"},"id":"did:example:ebfeb1f712ebc6f1c276e12ec21","name":"Jayden Doe","spouse":"did:example:c276e12ec21ebfeb1f712ebc6f1"},"expirationDate":"2020-01-01T19:23:24Z","id":"http://example.edu/credentials/1872","issuanceDate":"2010-01-01T19:23:24Z","issuer":{"id":"did:example:76e12ec712ebc6f1c221ebfeb1f","name":"Example University"},"referenceNumber":83294847,"type":["VerifiableCredential","UniversityDegreeCredential"]}
}

//nolint:lll
func ExampleCredential_JWTClaims() {
	// The Holder kept the credential serialized to JSON in her personal verifiable credential wallet.
	vcStrFromWallet := `
{
  "@context": [
    "https://www.w3.org/2018/credentials/v1",
    "https://www.w3.org/2018/credentials/examples/v1"
  ],
  "credentialSchema": [],
  "credentialSubject": {
    "degree": {
      "type": "BachelorDegree",
      "university": "MIT"
    },
    "id": "did:example:ebfeb1f712ebc6f1c276e12ec21",
    "name": "Jayden Doe",
    "spouse": "did:example:c276e12ec21ebfeb1f712ebc6f1"
  },
  "expirationDate": "2020-01-01T19:23:24Z",
  "id": "http://example.edu/credentials/1872",
  "issuanceDate": "2010-01-01T19:23:24Z",
  "issuer": {
    "id": "did:example:76e12ec712ebc6f1c221ebfeb1f",
    "name": "Example University"
  },
  "referenceNumber": 83294847,
  "type": [
    "VerifiableCredential",
    "UniversityDegreeCredential"
  ]
}
`

	// The Holder wants to send the credential to the Verifier in JWS.
	vc, _, err := verifiable.NewCredential([]byte(vcStrFromWallet))
	if err != nil {
		fmt.Println(fmt.Errorf("failed to decode VC JSON: %w", err))
	}

	jwtClaims, err := vc.JWTClaims(true)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to marshal JWT claims of VC: %w", err))
	}

	jws, err := jwtClaims.MarshalJWS(verifiable.EdDSA, getSigner(issuerPrivKey), "")
	if err != nil {
		fmt.Println(fmt.Errorf("failed to sign VC inside JWT: %w", err))
	}

	// The Holder passes JWS to Verifier
	fmt.Println(jws)

	// Output: eyJhbGciOiJFZERTQSIsImtpZCI6IiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1Nzc5MDY2MDQsImlhdCI6MTI2MjM3MzgwNCwiaXNzIjoiZGlkOmV4YW1wbGU6NzZlMTJlYzcxMmViYzZmMWMyMjFlYmZlYjFmIiwianRpIjoiaHR0cDovL2V4YW1wbGUuZWR1L2NyZWRlbnRpYWxzLzE4NzIiLCJuYmYiOjEyNjIzNzM4MDQsInN1YiI6ImRpZDpleGFtcGxlOmViZmViMWY3MTJlYmM2ZjFjMjc2ZTEyZWMyMSIsInZjIjp7IkBjb250ZXh0IjpbImh0dHBzOi8vd3d3LnczLm9yZy8yMDE4L2NyZWRlbnRpYWxzL3YxIiwiaHR0cHM6Ly93d3cudzMub3JnLzIwMTgvY3JlZGVudGlhbHMvZXhhbXBsZXMvdjEiXSwiY3JlZGVudGlhbFNjaGVtYSI6W10sImNyZWRlbnRpYWxTdWJqZWN0Ijp7ImRlZ3JlZSI6eyJ0eXBlIjoiQmFjaGVsb3JEZWdyZWUiLCJ1bml2ZXJzaXR5IjoiTUlUIn0sImlkIjoiZGlkOmV4YW1wbGU6ZWJmZWIxZjcxMmViYzZmMWMyNzZlMTJlYzIxIiwibmFtZSI6IkpheWRlbiBEb2UiLCJzcG91c2UiOiJkaWQ6ZXhhbXBsZTpjMjc2ZTEyZWMyMWViZmViMWY3MTJlYmM2ZjEifSwiaXNzdWVyIjp7Im5hbWUiOiJFeGFtcGxlIFVuaXZlcnNpdHkifSwicmVmZXJlbmNlTnVtYmVyIjo4LjMyOTQ4NDdlKzA3LCJ0eXBlIjpbIlZlcmlmaWFibGVDcmVkZW50aWFsIiwiVW5pdmVyc2l0eURlZ3JlZUNyZWRlbnRpYWwiXX19.IbGTRreDq5wZj3dETz6pnkRHdZOPwm9X-yQ-zyEve7bhOrZucqXRjmCLMkVFjNfcjJy5Xu4f3bcQe71pM--QCA
}

//nolint:lll
func ExampleCredential_AddLinkedDataProof() {
	vcJSON := `
{
  "@context": [
    "https://www.w3.org/2018/credentials/v1",
    "https://www.w3.org/2018/credentials/examples/v1"
  ],
  "credentialSchema": [],
  "credentialSubject": {
    "degree": {
      "type": "BachelorDegree",
      "university": "MIT"
    },
    "id": "did:example:ebfeb1f712ebc6f1c276e12ec21",
    "name": "Jayden Doe",
    "spouse": "did:example:c276e12ec21ebfeb1f712ebc6f1"
  },
  "expirationDate": "2020-01-01T19:23:24Z",
  "id": "http://example.edu/credentials/1872",
  "issuanceDate": "2009-01-01T19:23:24Z",
  "issuer": {
    "id": "did:example:76e12ec712ebc6f1c221ebfeb1f",
    "name": "Example University"
  },
  "referenceNumber": 83294849,
  "type": [
    "VerifiableCredential",
    "UniversityDegreeCredential"
  ]
}
`

	vc, _, err := verifiable.NewCredential([]byte(vcJSON))
	if err != nil {
		fmt.Println(fmt.Errorf("failed to decode VC JSON: %w", err))
	}

	err = vc.AddLinkedDataProof(&verifiable.LinkedDataProofContext{
		Created:                 &issued,
		SignatureType:           "Ed25519Signature2018",
		Suite:                   ed25519signature2018.New(ed25519signature2018.WithSigner(getSigner(issuerPrivKey))),
		SignatureRepresentation: verifiable.SignatureJWS,
		VerificationMethod:      "did:example:123456#key1",
	})
	if err != nil {
		fmt.Println(fmt.Errorf("failed to add linked data proof: %w", err))
	}

	vcJSONWithProof, err := json.MarshalIndent(vc, "", "\t")
	if err != nil {
		fmt.Println(fmt.Errorf("failed to marshal VC to JSON: %w", err))
	}

	fmt.Println(string(vcJSONWithProof))

	// Output: {
	//	"@context": [
	//		"https://www.w3.org/2018/credentials/v1",
	//		"https://www.w3.org/2018/credentials/examples/v1"
	//	],
	//	"credentialSchema": [],
	//	"credentialSubject": {
	//		"degree": {
	//			"type": "BachelorDegree",
	//			"university": "MIT"
	//		},
	//		"id": "did:example:ebfeb1f712ebc6f1c276e12ec21",
	//		"name": "Jayden Doe",
	//		"spouse": "did:example:c276e12ec21ebfeb1f712ebc6f1"
	//	},
	//	"expirationDate": "2020-01-01T19:23:24Z",
	//	"id": "http://example.edu/credentials/1872",
	//	"issuanceDate": "2009-01-01T19:23:24Z",
	//	"issuer": {
	//		"id": "did:example:76e12ec712ebc6f1c221ebfeb1f",
	//		"name": "Example University"
	//	},
	//	"proof": {
	//		"created": "2010-01-01T19:23:24Z",
	//		"jws": "eyJhbGciOiJFZDI1NTE5U2lnbmF0dXJlMjAxOCIsImI2NCI6ZmFsc2UsImNyaXQiOlsiYjY0Il19..Io85NnajfPXWBtB60QRI-OEfJtKEVv_ij2QTVLYqXdHTs01zCVMbUTyi6m5zIfH6YZELPgty2NujKYlun4L8Dg",
	//		"type": "Ed25519Signature2018",
	//		"verificationMethod": "did:example:123456#key1"
	//	},
	//	"referenceNumber": 83294849,
	//	"type": [
	//		"VerifiableCredential",
	//		"UniversityDegreeCredential"
	//	]
	//}
}

func getSigner(privKey []byte) *testSigner {
	return &testSigner{privateKey: privKey}
}

type testSigner struct {
	privateKey []byte
}

func (s *testSigner) Sign(doc []byte) ([]byte, error) {
	if l := len(s.privateKey); l != ed25519.PrivateKeySize {
		return nil, errors.New("ed25519: bad private key length")
	}

	return ed25519.Sign(s.privateKey, doc), nil
}
