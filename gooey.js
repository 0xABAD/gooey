(function () {
    const CONNECTING = 0;
    const OPEN       = 1;
    const CLOSING    = 2;
    const CLOSED     = 3;
    const SECONDS    = 1000;

    let location = window.location.href.substring(7);
    let socket   = new WebSocket('ws://' + location + 'gooeywebsocket');
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

    socket.addEventListener('message', function(wsevt) {
        let data     = JSON.parse(wsevt.data);
        let doReload = (data.hasOwnProperty('GooeyMessage') &&
                        data.hasOwnProperty('GooeyContent') &&
                        data.GooeyMessage === 'gooey-server-reload-content');

        if (doReload) {
            console.clear();
            console.log("[GOOEY] Hot reload.");

            let cnt = data.GooeyContent;

            if (cnt.Body !== "") {
                document.body.innerHTML = cnt.Body;
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
                // Unlike a style tag, we can't just replace the inner HTML
                // of the current script tag and have it reload.  Instead,
                // yank it out of the DOM and put it back in.
                script = document.createElement('script');
                script.id = "gooey-reload-js-content";
                script.innerHTML = cnt.Javascript;
                document.head.appendChild(script);
            }
        } else {
            gooey.OnMessage(data);
        }
    });
})();
