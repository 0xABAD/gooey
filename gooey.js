(function () {
    const CONNECTING = 0;
    const OPEN       = 1;
    const CLOSING    = 2;
    const CLOSED     = 3;

    let location = window.location.origin.substring(7);
    let socket   = new WebSocket('ws://' + location + '/gooeywebsocket');
    let gooey    = undefined;

    // Refer to gooey instead of window.gooey for better minification.
    if (window.hasOwnProperty("gooey")) {
        gooey = window.gooey;
    } else {
        gooey = {};
        window.gooey = gooey;

        gooey.OnMessage = function(msg) { console.log(msg); };
        gooey.Send = function(payload) {
            if (socket.readyState === OPEN) {
                socket.send(JSON.stringify(payload));
            } else {
                console.error('[GOOEY] Websocket connection is not open.');
            }
        };
        gooey.IsDisconnected = false;
        gooey.OnOpen = function() {
            console.log('[GOOEY] Websocket connection is open.');
        };
        gooey.OnDisconnect = function() {
            console.error('[GOOEY] Disconnected from server.');
        };
        gooey.OpenNewTab = function() {
            let req = new XMLHttpRequest();
            req.open('GET', window.location + 'gooeynewtab', true);
            req.send();
        };
    }

    let timeoutID = window.setInterval(function () {
        if (socket.readyState === CLOSED) {
            window.clearInterval(timeoutID);
            gooey.IsDisconnected = true;
            gooey.OnDisconnect();
        }
    }, 1500);

    socket.addEventListener('open', function() {
        gooey.IsDisconnected = false;
        gooey.OnOpen();
    });

    socket.addEventListener('message', function(wsevt) {
        let data     = JSON.parse(wsevt.data);
        let doReload = (data.hasOwnProperty('GooeyMessage') &&
                        data.hasOwnProperty('GooeyContent') &&
                        data.GooeyMessage === 'gooey-server-reload-content');

        if (doReload) {
            let cnt = data.GooeyContent;

            function replaceJS(js) {
                // Unlike a style tag, we can't just replace the inner HTML
                // of the current script tag and have it reload.  Instead,
                // yank it out of the DOM and put it back in.
                let s = document.createElement('script');
                s.id = "gooey-reload-js-content";
                s.innerHTML = js;
                document.head.appendChild(s);
            }

            // The style and script calls have to be duplicated because for
            // some odd reason, the 'if (script) {' statement below would
            // report an error of not being defined!
            if (cnt.Body !== "") {
                document.body.innerHTML = cnt.Body;
                let script = document.getElementById("gooey-reload-js-content");
                if (script) {
                    document.head.removeChild(script);
                    replaceJS(script.innerHTML);
                }
            }
            if (cnt.CSS !== "") {
                let style = document.getElementById("gooey-reload-css-content");
                if (style) {
                    style.innerHTML = cnt.CSS;
                } else {
                    style = document.createElement('style');
                    style.id = "gooey-reload-css-content";
                    style.innerHTML = cnt.CSS;
                    document.head.appendChild(style);
                }
            }
            if (cnt.Javascript !== "") {
                let script = document.getElementById("gooey-reload-js-content");
                if (script) {
                    document.head.removeChild(script);
                }
                replaceJS(cnt.Javascript);
            }
        } else {
            gooey.OnMessage(data);
        }
    });
})();
