// Code generated by qtc from "settings.qtpl". DO NOT EDIT.
// See https://github.com/valyala/quicktemplate for details.

//line transports/webui/views/settings.qtpl:1
package views

//line transports/webui/views/settings.qtpl:1
import (
	ory "github.com/ory/client-go"
)

//line transports/webui/views/settings.qtpl:5
import (
	qtio422016 "io"

	qt422016 "github.com/valyala/quicktemplate"
)

//line transports/webui/views/settings.qtpl:5
var (
	_ = qtio422016.Copy
	_ = qt422016.AcquireByteBuffer
)

//line transports/webui/views/settings.qtpl:6
type SettingsPage struct {
	Base
	Flow *ory.SettingsFlow
}

//line transports/webui/views/settings.qtpl:11
func (p SettingsPage) StreamBody(qw422016 *qt422016.Writer) {
//line transports/webui/views/settings.qtpl:11
	qw422016.N().S(`
        <section id="apiKeyContent" class="hidden px-6 mb-4 space-y-2"></section>
        <section id="mainContent" class="px-6 mb-4 space-y-2">
            <!-- <h1 class="text-center text-xl m-6">Account Settings</h1> -->
            `)
//line transports/webui/views/settings.qtpl:15
	p.Base.FlashMessagesContainer.StreamRender(qw422016)
//line transports/webui/views/settings.qtpl:15
	qw422016.N().S(`
            <span class="block text-center">`)
//line transports/webui/views/settings.qtpl:16
	streamrenderMessage(qw422016, p.Flow.Ui)
//line transports/webui/views/settings.qtpl:16
	qw422016.N().S(`</span>
            <div class="space-y-8">
                <section>
                    <h2 class="text-lg" id="profile">Change profile</h2>
                    `)
//line transports/webui/views/settings.qtpl:20
	streamformStart(qw422016, p.Flow.Ui)
//line transports/webui/views/settings.qtpl:20
	qw422016.N().S(`
                        `)
//line transports/webui/views/settings.qtpl:21
	streamrenderGroup(qw422016, p.Flow.Ui, "profile")
//line transports/webui/views/settings.qtpl:21
	qw422016.N().S(`
                        `)
//line transports/webui/views/settings.qtpl:22
	streamrenderSubmit(qw422016, p.Flow.Ui, "profile")
//line transports/webui/views/settings.qtpl:22
	qw422016.N().S(`
                    `)
//line transports/webui/views/settings.qtpl:23
	streamformEnd(qw422016)
//line transports/webui/views/settings.qtpl:23
	qw422016.N().S(`
                </section>
                <hr>
                <section>
                    <h2 class="text-lg" id="password">Change password</h2>
                    `)
//line transports/webui/views/settings.qtpl:28
	streamformStart(qw422016, p.Flow.Ui)
//line transports/webui/views/settings.qtpl:28
	qw422016.N().S(`
                        `)
//line transports/webui/views/settings.qtpl:29
	streamrenderGroup(qw422016, p.Flow.Ui, "password")
//line transports/webui/views/settings.qtpl:29
	qw422016.N().S(`
                        `)
//line transports/webui/views/settings.qtpl:30
	streamrenderSubmit(qw422016, p.Flow.Ui, "password")
//line transports/webui/views/settings.qtpl:30
	qw422016.N().S(`
                    `)
//line transports/webui/views/settings.qtpl:31
	streamformEnd(qw422016)
//line transports/webui/views/settings.qtpl:31
	qw422016.N().S(`
                </section>
                `)
//line transports/webui/views/settings.qtpl:33
	if hasGroup(p.Flow.Ui, "lookup_secret") {
//line transports/webui/views/settings.qtpl:33
		qw422016.N().S(`
                <hr>
                <section>
                    <h2 class="text-lg" id="backupcodes">2FA Backup Codes</h2>
                    `)
//line transports/webui/views/settings.qtpl:37
		streamformStart(qw422016, p.Flow.Ui)
//line transports/webui/views/settings.qtpl:37
		qw422016.N().S(`
                        `)
//line transports/webui/views/settings.qtpl:38
		streamrenderGroup(qw422016, p.Flow.Ui, "lookup_secret")
//line transports/webui/views/settings.qtpl:38
		qw422016.N().S(`
                        `)
//line transports/webui/views/settings.qtpl:39
		streamrenderSubmit(qw422016, p.Flow.Ui, "lookup_secret")
//line transports/webui/views/settings.qtpl:39
		qw422016.N().S(`
                    `)
//line transports/webui/views/settings.qtpl:40
		streamformEnd(qw422016)
//line transports/webui/views/settings.qtpl:40
		qw422016.N().S(`
                </section>
                `)
//line transports/webui/views/settings.qtpl:42
	}
//line transports/webui/views/settings.qtpl:42
	qw422016.N().S(`
                `)
//line transports/webui/views/settings.qtpl:43
	if hasGroup(p.Flow.Ui, "totp") {
//line transports/webui/views/settings.qtpl:43
		qw422016.N().S(`
                <hr>
                <section>
                    <h2 class="text-lg" id="2fa">2FA Authenticator App</h2>
                    `)
//line transports/webui/views/settings.qtpl:47
		streamformStart(qw422016, p.Flow.Ui)
//line transports/webui/views/settings.qtpl:47
		qw422016.N().S(`
                        `)
//line transports/webui/views/settings.qtpl:48
		streamrenderGroup(qw422016, p.Flow.Ui, "totp")
//line transports/webui/views/settings.qtpl:48
		qw422016.N().S(`
                        `)
//line transports/webui/views/settings.qtpl:49
		streamrenderSubmit(qw422016, p.Flow.Ui, "totp")
//line transports/webui/views/settings.qtpl:49
		qw422016.N().S(`
                    `)
//line transports/webui/views/settings.qtpl:50
		streamformEnd(qw422016)
//line transports/webui/views/settings.qtpl:50
		qw422016.N().S(`
                </section>
                `)
//line transports/webui/views/settings.qtpl:52
	}
//line transports/webui/views/settings.qtpl:52
	qw422016.N().S(`
                `)
//line transports/webui/views/settings.qtpl:53
	if hasGroup(p.Flow.Ui, "webauthn") {
//line transports/webui/views/settings.qtpl:53
		qw422016.N().S(`
                <hr>
                <section>
                    <h2 class="text-lg" id="webauthn">Web Authentication</h2>
                    `)
//line transports/webui/views/settings.qtpl:57
		streamformStart(qw422016, p.Flow.Ui)
//line transports/webui/views/settings.qtpl:57
		qw422016.N().S(`
                        `)
//line transports/webui/views/settings.qtpl:58
		streamrenderGroup(qw422016, p.Flow.Ui, "webauthn")
//line transports/webui/views/settings.qtpl:58
		qw422016.N().S(`
                        `)
//line transports/webui/views/settings.qtpl:59
		streamrenderSubmit(qw422016, p.Flow.Ui, "webauthn")
//line transports/webui/views/settings.qtpl:59
		qw422016.N().S(`
                    `)
//line transports/webui/views/settings.qtpl:60
		streamformEnd(qw422016)
//line transports/webui/views/settings.qtpl:60
		qw422016.N().S(`
                </section>
                `)
//line transports/webui/views/settings.qtpl:62
	}
//line transports/webui/views/settings.qtpl:62
	qw422016.N().S(`
            </section>
`)
//line transports/webui/views/settings.qtpl:64
}

//line transports/webui/views/settings.qtpl:64
func (p SettingsPage) WriteBody(qq422016 qtio422016.Writer) {
//line transports/webui/views/settings.qtpl:64
	qw422016 := qt422016.AcquireWriter(qq422016)
//line transports/webui/views/settings.qtpl:64
	p.StreamBody(qw422016)
//line transports/webui/views/settings.qtpl:64
	qt422016.ReleaseWriter(qw422016)
//line transports/webui/views/settings.qtpl:64
}

//line transports/webui/views/settings.qtpl:64
func (p SettingsPage) Body() string {
//line transports/webui/views/settings.qtpl:64
	qb422016 := qt422016.AcquireByteBuffer()
//line transports/webui/views/settings.qtpl:64
	p.WriteBody(qb422016)
//line transports/webui/views/settings.qtpl:64
	qs422016 := string(qb422016.B)
//line transports/webui/views/settings.qtpl:64
	qt422016.ReleaseByteBuffer(qb422016)
//line transports/webui/views/settings.qtpl:64
	return qs422016
//line transports/webui/views/settings.qtpl:64
}
