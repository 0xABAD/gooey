package gooey

const INDEX = `
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>Gooey App</title>
<script>
(function(){var e=window.location.origin.substring(7),d=new WebSocket("ws://"+e+"/gooeywebsocket"),a=void 0;window.hasOwnProperty("gooey")?a=window.gooey:(a={},window.gooey=a,a.OnMessage=function(a){console.log(a)},a.Send=function(a){1===d.readyState?d.send(JSON.stringify(a)):console.error("[GOOEY] Websocket connection is not open.")},a.IsDisconnected=!1,a.OnOpen=function(){console.log("[GOOEY] Websocket connection is open.")},a.OnDisconnect=function(){console.error("[GOOEY] Disconnected from server.")},
a.OpenNewTab=function(){var a=new XMLHttpRequest;a.open("GET",window.location+"gooeynewtab",!0);a.send()});var f=window.setInterval(function(){3===d.readyState&&(window.clearInterval(f),a.IsDisconnected=!0,a.OnDisconnect())},1500);d.addEventListener("open",function(){a.IsDisconnected=!1;a.OnOpen()});d.addEventListener("message",function(d){var b=JSON.parse(d.data);if(b.hasOwnProperty("GooeyMessage")&&b.hasOwnProperty("GooeyContent")&&"gooey-server-reload-content"===b.GooeyMessage){d=function(a){var b=
document.createElement("script");b.id="gooey-reload-js-content";b.innerHTML=a;document.head.appendChild(b)};b=b.GooeyContent;if(""!==b.Body){document.body.innerHTML=b.Body;var c=document.getElementById("gooey-reload-js-content");c&&(document.head.removeChild(c),d(c.innerHTML))}""!==b.CSS&&((c=document.getElementById("gooey-reload-css-content"))?c.innerHTML=b.CSS:(c=document.createElement("style"),c.id="gooey-reload-css-content",c.innerHTML=b.CSS,document.head.appendChild(c)));""!==b.Javascript&&((c=
document.getElementById("gooey-reload-js-content"))&&document.head.removeChild(c),d(b.Javascript))}else a.OnMessage(b)})})();
</script>
<script>
(function() {
window.gooey.OnMessage = function(msg) {
var elt = document.getElementById('gooey-message-area');
elt.innerText = msg;
};
})();
</script>
</head>
<body>
<h1>Gooey Sample Index Page</h1>
<div>
<h3>Last Received Gooey Message:</h3>
<div id="gooey-message-area"></div>
</div>
</body>
</html>
`
