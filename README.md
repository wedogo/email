# email
[![Build Status](https://travis-ci.org/wedogo/email.svg?branch=develop)](https://travis-ci.org/wedogo/email)
[![Join the chat at https://gitter.im/wedogo/email](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/wedogo/email?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
A GO library for the creation of email messages.

## Design philosofies
Normal action should be easy and somewhat fool-proof, advanced actions must be 
possible, but are not necessary fool-proof. As an example, the format of the 
"to" header is strictly checked, but not all (obscure) headers are checked if 
they conform the rfc5322

### Standards
 - Following RFC standards where possible
 - Following common practices when in line with RFC's
 - Following common practices when not line with RFC's only when necessary

### Code
 - Targeting Go 1.2 or later
 - No dependencies outside Go standard lib
 - Fully tested
 - Following the Go style guide and common practices
 - Possibly in line with parsing packages in net/mail and mime

### Scope
 - Only creation of messages and exporting them to bytes or writer
 - Sending (with smtp or other method) is explicitly not in scope. It might come later in another (sub)-package
 
### Target example cases
 - Send a text email to the linux kernel mailing list with patch as inline attachment
 - Send a registration email in both html and text
 - Send an email with pictures "inline"
 - Send an email with pictures as attachments

# Other Go email packages
This is a list of other Go packages that aim to do similar things. Some of these are used for inspiration. Any critism mentioned with the package is just meant as a warning to self to avoid similar pitfalls.

<dl>
<dt>https://github.com/jpoehls/gophermail</dt>
<dd>Does not wrap subject header. Contains unnecessary parts in body</dd>
<dt>https://github.com/go-gomail/gomail</dt>
<dd>Does not export complete message directly (tight integration with sending mail). Requires Go 1.3</dd>
<dt>https://github.com/jordan-wright/email</dt>
<dd>Does not wrap or encode subject header</dd>
</dl>
