// Copyright (c) 2026 hangtiancheng
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package code

type Code int

const (
	OK               Code = 1000
	ParamsInvalid    Code = 2001
	UserExist        Code = 2002
	UserNotExist     Code = 2003
	PasswordError    Code = 2004
	PasswordNotMatch Code = 2005
	TokenInvalid     Code = 2006
	NotLogin         Code = 2007
	CaptchaInvalid   Code = 2008
	RecordNotFound   Code = 2009
	PasswordIllegal  Code = 2010
	Forbidden        Code = 3001
	ServerError      Code = 4001
	ModelNotFound    Code = 5001
	ModelNoPerm      Code = 5002
	ModelError       Code = 5003
)

var messages = map[Code]string{
	OK:               "OK",
	ParamsInvalid:    "Params Invalid",
	UserExist:        "User Exist",
	UserNotExist:     "User Not Exist",
	PasswordError:    "Password Error",
	PasswordNotMatch: "Password Not Match",
	TokenInvalid:     "Token Invalid",
	NotLogin:         "Not Login",
	CaptchaInvalid:   "Captcha Invalid",
	RecordNotFound:   "Record Not Found",
	PasswordIllegal:  "Password Illegal",
	Forbidden:        "Forbidden",
	ServerError:      "Server Error",
	ModelNotFound:    "Model Not Found",
	ModelNoPerm:      "Model No Permission",
	ModelError:       "Model Error",
}

func Message(c Code) string {
	if msg, ok := messages[c]; ok {
		return msg
	}
	return messages[ServerError]
}
