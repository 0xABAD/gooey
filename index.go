package gooey

const INDEX = `
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>Gooey App</title>
<script>
(function(){var e=window.location.href.substring(7),d=new WebSocket("ws://"+e+"gooeywebsocket"),b=void 0;window.hasOwnProperty("gooey")?b=window.gooey:(b={},window.gooey=b,b.OnMessage=function(a){console.log(a)},b.Send=function(a){1===d.readyState?d.send(JSON.stringify(a)):console.error("[GOOEY] Websocket connection is not open.")},b.IsDisconnected=!1,b.OnDisconnect=function(){console.error("[GOOEY] Disconnected from server.")},b.OpenNewTab=function(){var a=new XMLHttpRequest;a.open("GET",window.location+
"gooeynewtab",!0);a.send()});var f=window.setInterval(function(){3===d.readyState&&(window.clearInterval(f),b.IsDisconnected=!0,b.OnDisconnect())},1500);d.addEventListener("message",function(a){function d(a){var b=document.createElement("script");b.id="gooey-reload-js-content";b.innerHTML=a;document.head.appendChild(b)}a=JSON.parse(a.data);if(a.hasOwnProperty("GooeyMessage")&&a.hasOwnProperty("GooeyContent")&&"gooey-server-reload-content"===a.GooeyMessage){a=a.GooeyContent;if(""!==a.Body){document.body.innerHTML=
a.Body;var c=document.getElementById("gooey-reload-js-content");c&&(document.head.removeChild(c),d(c.innerHTML))}""!==a.CSS&&((c=document.getElementById("gooey-reload-css-content"))?c.innerHTML=a.CSS:(c=document.createElement("style"),c.id="gooey-reload-css-content",c.innerHTML=a.CSS,document.head.appendChild(c)));""!==a.Javascript&&((c=document.getElementById("gooey-reload-js-content"))&&document.head.removeChild(c),d(a.Javascript))}else b.OnMessage(a)})})();
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
<div>
</body>
</html>
`
