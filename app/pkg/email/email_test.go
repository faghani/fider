package email_test

import (
	"testing"

	"github.com/getfider/fider/app/pkg/log/noop"
	"github.com/getfider/fider/app/pkg/worker"

	"github.com/getfider/fider/app/pkg/email"

	. "github.com/getfider/fider/app/pkg/assert"
)

func TestRenderMessage(t *testing.T) {
	RegisterT(t)

	ctx := worker.NewContext("ID-1", "TaskName", nil, noop.NewLogger())
	message := email.RenderMessage(ctx, "echo_test", email.Params{
		"name": "Fider",
	})
	Expect(message.Subject).Equals("Message to: Fider")
	Expect(message.Body).Equals(`<!DOCTYPE html PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN" "http://www.w3.org/TR/html4/loose.dtd">
<html>
	<head>
		<meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
		<meta name="viewport" content="width=device-width">
		<meta http-equiv="X-UA-Compatible" content="IE=edge">
	</head>
	<body bgcolor="#F7F7F7" style="font-size:16px">
		<table width="100%" bgcolor="#F7F7F7" cellpadding="0" cellspacing="0" border="0" style="text-align:center;font-size:14px;">
			<tr>
				<td height="40">&nbsp;</td>
			</tr>
			
			<tr>
				<td align="center">
					<table bgcolor="#FFFFFF" cellpadding="0" cellspacing="0" border="0" style="text-align:left;padding:20px;margin:10px;border-radius:5px;color:#1c262d;border:1px solid #ECECEC;min-width:320px;max-width:660px;">
						Hello World Fider!
					</table>
				</td>
			</tr>
			<tr>
				<td>
					<span style="color:#666;font-size:11px">This email was sent from a notification-only address that cannot accept incoming email. Please do not reply to this message.</span>
				</td>
			</tr>
			<tr>
				<td height="40">&nbsp;</td>
			</tr>
		</table>
	</body>
</html>`)
}

func TestCanSendTo(t *testing.T) {
	RegisterT(t)

	testCases := []struct {
		whitelist string
		blacklist string
		input     []string
		canSend   bool
	}{
		{
			whitelist: "(^.+@fider.io$)|(^darthvader\\.fider(\\+.*)?@gmail\\.com$)",
			blacklist: "",
			input:     []string{"me@fider.io", "me+123@fider.io", "darthvader.fider@gmail.com", "darthvader.fider+434@gmail.com"},
			canSend:   true,
		},
		{
			whitelist: "(^.+@fider.io$)|(^darthvader\\.fider(\\+.*)?@gmail\\.com$)",
			blacklist: "",
			input:     []string{"me+123@fider.iod", "me@fidero.io", "darthvader.fidera@gmail.com", "@fider.io"},
			canSend:   false,
		},
		{
			whitelist: "(^.+@fider.io$)|(^darthvader\\.fider(\\+.*)?@gmail\\.com$)",
			blacklist: "(^.+@fider.io$)",
			input:     []string{"me@fider.io"},
			canSend:   true,
		},
		{
			whitelist: "",
			blacklist: "(^.+@fider.io$)",
			input:     []string{"me@fider.io", "abc@fider.io"},
			canSend:   false,
		},
		{
			whitelist: "",
			blacklist: "(^.+@fider.io$)",
			input:     []string{"me@fider.com", "abc@fiderio.io"},
			canSend:   true,
		},
		{
			whitelist: "",
			blacklist: "",
			input:     []string{"me@fider.io"},
			canSend:   true,
		},
		{
			whitelist: "",
			blacklist: "",
			input:     []string{"", " "},
			canSend:   false,
		},
	}

	for _, testCase := range testCases {
		email.SetWhitelist(testCase.whitelist)
		email.SetBlacklist(testCase.blacklist)
		for _, input := range testCase.input {
			Expect(email.CanSendTo(input)).Equals(testCase.canSend)
		}
	}
}

func TestParamsMerge(t *testing.T) {
	RegisterT(t)

	p1 := email.Params{
		"name": "Jon",
		"age":  26,
	}
	p2 := p1.Merge(email.Params{
		"age":   30,
		"email": "john.snow@got.com",
	})
	Expect(p2).Equals(email.Params{
		"name":  "Jon",
		"age":   30,
		"email": "john.snow@got.com",
	})
}
