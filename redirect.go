package gooey

const REDIRECT = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>(function(){window.location="http://{{.Addr}}";})()</script>
</head>
<body></body>
</html>`
