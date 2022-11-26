// Package aws provides the functionality about AWS.
package aws

import (
	"bytes"
	"compress/flate"
	"context"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

const (
	// EndpointURL receives SAML response.
	EndpointURL = "https://signin.aws.amazon.com/saml"

	roleAttributeName = "https://aws.amazon.com/SAML/Attributes/Role"
)

// SAMLResponse is SAML response
type SAMLResponse struct {
	Assertion Assertion
}

// Assertion is an Assertion element of SAML response
type Assertion struct {
	AttributeStatement AttributeStatement
}

// AttributeStatement is an AttributeStatement element of SAML response
type AttributeStatement struct {
	Attributes []Attribute `xml:"Attribute"`
}

// Attribute is an Attribute element of SAML response
type Attribute struct {
	Name            string           `xml:",attr"`
	AttributeValues []AttributeValue `xml:"AttributeValue"`
}

// AttributeValue is an AttributeValue element of SAML response
type AttributeValue struct {
	Value string `xml:",innerxml"`
}

// ParseSAMLResponse parses base64 encoded response to SAMLResponse structure
func ParseSAMLResponse(base64Response string) (*SAMLResponse, error) {
	responseData, err := base64.StdEncoding.DecodeString(base64Response)
	if err != nil {
		return nil, err
	}

	response := SAMLResponse{}
	err = xml.Unmarshal(responseData, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// ExtractRoleArnAndPrincipalArn extracts role ARN and principal ARN from SAML response
func ExtractRoleArnAndPrincipalArn(samlResponse SAMLResponse, roleName string, accountId string) (string, string, error) {
	for _, attr := range samlResponse.Assertion.AttributeStatement.Attributes {
		if attr.Name != roleAttributeName {
			continue
		}

		for _, v := range attr.AttributeValues {
			s := strings.Split(v.Value, ",")
			roleArn := s[0]
			principalArn := s[1]
			if roleName != "" && strings.Split(roleArn, "/")[1] != roleName || !strings.Contains(roleArn, accountId) {
				continue
			}
			return roleArn, principalArn, nil
		}
	}

	return "", "", fmt.Errorf("no such attribute: %s", roleAttributeName)
}

// AssumeRoleWithSAML sends a AssumeRoleWithSAML request to AWS and returns credentials
func AssumeRoleWithSAML(ctx context.Context, durationHours int, roleArn string, principalArn string, base64Response string) (*sts.Credentials, error) {
	sess := session.Must(session.NewSession())
	svc := sts.New(sess)

	input := sts.AssumeRoleWithSAMLInput{
		DurationSeconds: aws.Int64(int64(durationHours) * 60 * 60),
		RoleArn:         aws.String(roleArn),
		PrincipalArn:    aws.String(principalArn),
		SAMLAssertion:   aws.String(base64Response),
	}
	res, err := svc.AssumeRoleWithSAMLWithContext(ctx, &input)
	if err != nil {
		return nil, err
	}

	return res.Credentials, nil
}

func deflate(src string) (*bytes.Buffer, error) {
	b := new(bytes.Buffer)

	w, err := flate.NewWriter(b, 9)
	if err != nil {
		return nil, err
	}

	if _, err := w.Write([]byte(src)); err != nil {
		return nil, err
	}

	err = w.Close()
	if err != nil {
		return nil, err
	}

	return b, nil
}
