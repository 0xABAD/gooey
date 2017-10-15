gooey test
==========

Due to gooey's heavy IO nature and the client/server interaction, it's
hard to setup an automated test suite.  However, gooey does have a
small feature so here we just document a list of test cases to run
manually.  Each test is labeled by the file containing the test
and what is being tested.

Tests
-----

* **[devtest.go]** Modifying `web/body.html` causes hot reload to
activate. Any change should be reflected on the client browser window.

* **[devtest.go]** Modifying `web/body.html` causes `web/extra.js` to
be reloaded in the client window.  The console output from
`web/extra.js` should be seen when inspecting the client window.

* **[devtest.go]** Modifying `web/extra.js` causes hot reload to
activate.  Modifying the console.log statement should be reflected
when inspecting the client window.

* **[devtest.go]** Modifying `web/extra.css` causes hot reload to
activate.  Changing the background of the body should reflected in
the client window.

* **[devtest.go]** Adding a file that begins with `.#` to the web
directory should be ignored.  Create a `web/.#test.js` that prints
anything to the console and verify that the output is not seen.

* **[devtest.go]** Closing the last connected tab causes the server
to automatically shutdown.

* **[devtest.go]** Opening and then closing another tab does not cause
the server to automatically shutdown.
