tmpl = """<html>
<head>
<style type="text/css">

body, table {
	font-family: "Lucida Grande", sans-serif;
	font-size: 12px;
	font-size: 8pt;
}

table {
	color: #444;
}

td {
	font-family: consolas, menlo, monospace;
}

.header {
	color: #420066;
	color: #0000ff;
	font-style: italic;
}

.line {
	border-bottom: 1px dotted #ccc;
}

.big {
	font-size: 140%;
	font-weight: bold;
}

.comment {
	color: #999;
}

.em {
	font-weight: bold;
	color: #420066;
	color: #000;
	font-size: 130%;
	font-size: 100%;
}

</style>
</head>
<body>

<div>
    <a href="/index.html">home</a> &#8227;
	<a href="#number">Number</a> &bull;
	<a href="#string">String</a> &bull;
	<a href="#number-to-string">Number&lt;-&gt;String</a> &bull;
	<a href="#boolean">Boolean</a> &bull;
	<a href="#date">Date</a> &bull;
	<a href="#math">Math</a> &bull;
	<a href="#array">Array</a> &bull;
	<a href="#function">Function</a> &bull;
	<a href="#logic">logic</a> &bull;
	<a href="#object">Object</a> &bull;
	<a href="#type">type</a> &bull;
	<a href="#object-orientation">object-orientation</a> &bull;
	<a href="#exceptions">exceptions</a>
</div>
<br>
%s

<hr/>
<center><a href="/index.html">Krzysztof Kowalczyk</a></center>

<!-- Global site tag (gtag.js) - Google Analytics -->
<script async src="https://www.googletagmanager.com/gtag/js?id=UA-194516-1"></script>
<script>
  window.dataLayer = window.dataLayer || [];
  function gtag(){dataLayer.push(arguments);}
  gtag('js', new Date());

  gtag('config', 'UA-194516-1');
</script>

</body>
</html>"""
