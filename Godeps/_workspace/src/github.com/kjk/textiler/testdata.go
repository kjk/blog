package textiler

// Those tests are based on https://github.com/ikirudennis/python-textile/blob/master/textile/tests/__init__.py

var HtmlTests = []string{
	`I spoke.
And none replied.`,
	`	<p>I spoke.<br>
And none replied.</p>`,

	`I __know__.
I **really** __know__.`,
	`	<p>I <i>know</i>.<br>
I <b>really</b> <i>know</i>.</p>`,

	`I'm %{color:red}unaware%
of most soft drinks.`,
	`	<p>I&#8217;m <span style="color:red;">unaware</span><br>
of most soft drinks.</p>`,

	`I seriously *{color:red}blushed*
when I _(big)sprouted_ that
corn stalk from my
%[es]cabeza%.`,
	`	<p>I seriously <strong style="color:red;">blushed</strong><br>
when I <em class="big">sprouted</em> that<br>
corn stalk from my<br>
<span lang="es">cabeza</span>.</p>`,

	`<pre>
<code>
a.gsub!( /</, "" )
</code>
</pre>`,
	`<pre>
<code>
a.gsub!( /&lt;/, "" )
</code>
</pre>`,

	`<div style="float:right;">

h3. Sidebar

"Hobix":http://hobix.com/
"Ruby":http://ruby-lang.org/

</div>

The main text of the
page goes here and will
stay to the left of the
sidebar.`,
	`	<p><div style="float:right;"></p>

	<h3>Sidebar</h3>

	<p><a href="http://hobix.com/">Hobix</a><br>
<a href="http://ruby-lang.org/">Ruby</a></p>

	<p></div></p>

	<p>The main text of the<br>
page goes here and will<br>
stay to the left of the<br>
sidebar.</p>`,

	`I am crazy about "Hobix":hobix
and "it's":hobix "all":hobix I ever
"link to":hobix!

[hobix]http://hobix.com`,
	`	<p>I am crazy about <a href="http://hobix.com">Hobix</a><br>
and <a href="http://hobix.com">it&#8217;s</a> <a href="http://hobix.com">all</a> I ever<br>
<a href="http://hobix.com">link to</a>!</p>`,

	`!http://hobix.com/sample.jpg!`,
	`	<p><img src="http://hobix.com/sample.jpg" alt=""></p>`,

	`!openwindow1.gif(Bunny.)!`,
	`	<p><img src="openwindow1.gif" title="Bunny." alt="Bunny."></p>`,

	`!openwindow1.gif!:http://hobix.com/`,
	`	<p><a href="http://hobix.com/" class="img"><img src="openwindow1.gif" alt=""></a></p>`,

	`!>obake.gif!

And others sat all round the small
machine and paid it to sing to them.`,
	`	<p><img src="obake.gif" style="float: right;" alt=""></p>

	<p>And others sat all round the small<br>
machine and paid it to sing to them.</p>`,

	`!http://render.mathim.com/A%5EtAx%20%3D%20A%5Et%28Ax%29.!`,
	`	<p><img src="http://render.mathim.com/A%5EtAx%20%3D%20A%5Et%28Ax%29." alt=""></p>`,

	`notextile. <b> foo bar baz</b>

p. quux
`,
	`<b> foo bar baz</b>

	<p>quux</p>`,
}
var XhtmlTests = []string{
	`hello, world`,
	`	<p>hello, world</p>`,

	`A single paragraph.

Followed by another.`,
	`	<p>A single paragraph.</p>

	<p>Followed by another.</p>`,

	`I am <b>very</b> serious.

<pre>
I am <b>very</b> serious.
</pre>`,
	`	<p>I am <b>very</b> serious.</p>

<pre>
I am &lt;b&gt;very&lt;/b&gt; serious.
</pre>`,

	`I spoke.
And none replied.`,
	`	<p>I spoke.<br />
And none replied.</p>`,

	`"Observe!"`,
	`	<p>&#8220;Observe!&#8221; </p>`,

	`Observe -- very nice!`,
	`	<p>Observe &#8212; very nice!</p>`,

	`Observe - tiny and brief.`,
	`	<p>Observe &#8211; tiny and brief.</p>`,

	`Observe...`,
	`	<p>Observe&#8230;</p>`,

	`Observe ...`,
	`	<p>Observe &#8230;</p>`,

	`Observe: 2 x 2.`,
	`	<p>Observe: 2 &#215; 2.</p>`,

	`one(TM), two(R), three(C).`,
	`	<p>one&#8482;, two&#174;, three&#169;.</p>`,

	`h1. Header 1`,
	`	<h1>Header 1</h1>`,

	`h2. Header 2`,
	`	<h2>Header 2</h2>`,

	`h3. Header 3`,
	`	<h3>Header 3</h3>`,

	`An old text

bq. A block quotation.

Any old text`,
	`	<p>An old text</p>

	<blockquote>
		<p>A block quotation.</p>
	</blockquote>

	<p>Any old text</p>`,

	`I _believe_ every word.`,
	`	<p>I <em>believe</em> every word.</p>`,

	`And then? She *fell*!`,
	`	<p>And then? She <strong>fell</strong>!</p>`,

	`I __know__.
I **really** __know__.`,
	`	<p>I <i>know</i>.<br />
I <b>really</b> <i>know</i>.</p>`,

	`??Cat's Cradle?? by Vonnegut`,
	`	<p><cite>Cat&#8217;s Cradle</cite> by Vonnegut</p>`,

	`Convert with @str(foo)@`,
	`	<p>Convert with <code>str(foo)</code></p>`,

	`I'm -sure- not sure.`,
	`	<p>I&#8217;m <del>sure</del> not sure.</p>`,

	`You are a +pleasant+ child.`,
	`	<p>You are a <ins>pleasant</ins> child.</p>`,

	`a ^2^ + b ^2^ = c ^2^`,
	`	<p>a <sup>2</sup> + b <sup>2</sup> = c <sup>2</sup></p>`,

	`log ~2~ x`,
	`	<p>log <sub>2</sub> x</p>`,

	`I'm %unaware% of most soft drinks.`,
	`	<p>I&#8217;m <span>unaware</span> of most soft drinks.</p>`,

	`I'm %{color:red}unaware%
of most soft drinks.`,
	`	<p>I&#8217;m <span style="color:red;">unaware</span><br />
of most soft drinks.</p>`,

	`p(example1). An example`,
	`	<p class="example1">An example</p>`,

	`p(#big-red). Red here`,
	`	<p id="big-red">Red here</p>`,

	`p(example1#big-red2). Red here`,
	`	<p class="example1" id="big-red2">Red here</p>`,

	`p{color:blue;margin:30px}. Spacey blue`,
	`	<p style="color:blue; margin:30px;">Spacey blue</p>`,

	`p[fr]. rouge`,
	`	<p lang="fr">rouge</p>`,

	`I seriously *{color:red}blushed*
when I _(big)sprouted_ that
corn stalk from my
%[es]cabeza%.`,
	`	<p>I seriously <strong style="color:red;">blushed</strong><br />
when I <em class="big">sprouted</em> that<br />
corn stalk from my<br />
<span lang="es">cabeza</span>.</p>`,

	`p<. align left`,
	`	<p style="text-align:left;">align left</p>`,

	`p>. align right`,
	`	<p style="text-align:right;">align right</p>`,

	`p=. centered`,
	`	<p style="text-align:center;">centered</p>`,

	`p<>. justified`,
	`	<p style="text-align:justify;">justified</p>`,

	`p(. left ident 1em`,
	`	<p style="padding-left:1em;">left ident 1em</p>`,

	`p((. left ident 2em`,
	`	<p style="padding-left:2em;">left ident 2em</p>`,

	`p))). right ident 3em`,
	`	<p style="padding-right:3em;">right ident 3em</p>`,

	`h2()>. Bingo.`,
	`	<h2 style="padding-left:1em; padding-right:1em; text-align:right;">Bingo.</h2>`,

	`h3()>[no]{color:red}. Bingo`,
	`	<h3 style="color:red; padding-left:1em; padding-right:1em; text-align:right;" lang="no">Bingo</h3>`,

	`<pre>
<code>
a.gsub!( /</, "" )
</code>
</pre>`,
	`<pre>
<code>
a.gsub!( /&lt;/, "" )
</code>
</pre>`,

	`<div style="float:right;">

h3. Sidebar

"Hobix":http://hobix.com/
"Ruby":http://ruby-lang.org/

</div>

The main text of the
page goes here and will
stay to the left of the
sidebar.`,
	`	<p><div style="float:right;"></p>

	<h3>Sidebar</h3>

	<p><a href="http://hobix.com/">Hobix</a><br />
<a href="http://ruby-lang.org/">Ruby</a></p>

	<p></div></p>

	<p>The main text of the<br />
page goes here and will<br />
stay to the left of the<br />
sidebar.</p>`,

	`# A first item
# A second item
# A third`,
	`	<ol>
		<li>A first item</li>
		<li>A second item</li>
		<li>A third</li>
	</ol>`,

	`# Fuel could be:
## Coal
## Gasoline
## Electricity
# Humans need only:
## Water
## Protein`,
	`	<ol>
		<li>Fuel could be:
	<ol>
		<li>Coal</li>
		<li>Gasoline</li>
		<li>Electricity</li>
	</ol></li>
		<li>Humans need only:
	<ol>
		<li>Water</li>
		<li>Protein</li>
	</ol></li>
	</ol>`,

	`* A first item
* A second item
* A third`,
	`	<ul>
		<li>A first item</li>
		<li>A second item</li>
		<li>A third</li>
	</ul>`,

	`• A first item
• A second item
• A third`,
	`	<ul>
		<li>A first item</li>
		<li>A second item</li>
		<li>A third</li>
	</ul>`,

	`* Fuel could be:
** Coal
** Gasoline
** Electricity
* Humans need only:
** Water
** Protein`,
	`	<ul>
		<li>Fuel could be:
	<ul>
		<li>Coal</li>
		<li>Gasoline</li>
		<li>Electricity</li>
	</ul></li>
		<li>Humans need only:
	<ul>
		<li>Water</li>
		<li>Protein</li>
	</ul></li>
	</ul>`,

	`I searched "Google":http://google.com.`,
	`	<p>I searched <a href="http://google.com">Google</a>.</p>`,

	`I searched "a search engine (Google)":http://google.com.`,
	`	<p>I searched <a href="http://google.com" title="Google">a search engine</a>.</p>`,

	`I am crazy about "Hobix":hobix
and "it's":hobix "all":hobix I ever
"link to":hobix!

[hobix]http://hobix.com`,
	`	<p>I am crazy about <a href="http://hobix.com">Hobix</a><br />
and <a href="http://hobix.com">it&#8217;s</a> <a href="http://hobix.com">all</a> I ever<br />
<a href="http://hobix.com">link to</a>!</p>`,

	`!http://hobix.com/sample.jpg!`,
	`	<p><img src="http://hobix.com/sample.jpg" alt="" /></p>`,

	`!openwindow1.gif(Bunny.)!`,
	`	<p><img src="openwindow1.gif" title="Bunny." alt="Bunny." /></p>`,

	`!openwindow1.gif!:http://hobix.com/`,
	`	<p><a href="http://hobix.com/" class="img"><img src="openwindow1.gif" alt="" /></a></p>`,

	`!>obake.gif!

And others sat all round the small
machine and paid it to sing to them.`,
	`	<p><img src="obake.gif" style="float: right;" alt="" /></p>

	<p>And others sat all round the small<br />
machine and paid it to sing to them.</p>`,

	`We use CSS(Cascading Style Sheets).`,
	`	<p>We use <acronym title="Cascading Style Sheets"><span class="caps">CSS</span></acronym>.</p>`,

	`|one|two|three|
|a|b|c|`,
	`	<table>
		<tr>
			<td>one</td>
			<td>two</td>
			<td>three</td>
		</tr>
		<tr>
			<td>a</td>
			<td>b</td>
			<td>c</td>
		</tr>
	</table>`,

	`| name | age | sex |
| joan | 24 | f |
| archie | 29 | m |
| bella | 45 | f |`,
	`	<table>
		<tr>
			<td> name </td>
			<td> age </td>
			<td> sex </td>
		</tr>
		<tr>
			<td> joan </td>
			<td> 24 </td>
			<td> f </td>
		</tr>
		<tr>
			<td> archie </td>
			<td> 29 </td>
			<td> m </td>
		</tr>
		<tr>
			<td> bella </td>
			<td> 45 </td>
			<td> f </td>
		</tr>
	</table>`,

	`|_. name |_. age |_. sex |
| joan | 24 | f |
| archie | 29 | m |
| bella | 45 | f |`,
	`	<table>
		<tr>
			<th>name </th>
			<th>age </th>
			<th>sex </th>
		</tr>
		<tr>
			<td> joan </td>
			<td> 24 </td>
			<td> f </td>
		</tr>
		<tr>
			<td> archie </td>
			<td> 29 </td>
			<td> m </td>
		</tr>
		<tr>
			<td> bella </td>
			<td> 45 </td>
			<td> f </td>
		</tr>
	</table>`,

	`<script>alert("hello");</script>`,
	`	<p><script>alert(&#8220;hello&#8221;);</script></p>`,

	`pre.. Hello

Hello Again

p. normal text`,
	`<pre>Hello

Hello Again
</pre>

	<p>normal text</p>`,

	`<pre>this is in a pre tag</pre>`,
	`<pre>this is in a pre tag</pre>`,

	`"test1":http://foo.com/bar--baz

"test2":http://foo.com/bar---baz

"test3":http://foo.com/bar-17-18-baz`,
	`	<p><a href="http://foo.com/bar--baz">test1</a></p>

	<p><a href="http://foo.com/bar---baz">test2</a></p>

	<p><a href="http://foo.com/bar-17-18-baz">test3</a></p>`,

	`"foo ==(bar)==":#foobar`,
	`	<p><a href="#foobar">foo (bar)</a></p>`,

	`!http://render.mathim.com/A%5EtAx%20%3D%20A%5Et%28Ax%29.!`,
	`	<p><img src="http://render.mathim.com/A%5EtAx%20%3D%20A%5Et%28Ax%29." alt="" /></p>`,

	`* Point one
* Point two
## Step 1
## Step 2
## Step 3
* Point three
** Sub point 1
** Sub point 2`,
	`	<ul>
		<li>Point one</li>
		<li>Point two
	<ol>
		<li>Step 1</li>
		<li>Step 2</li>
		<li>Step 3</li>
	</ol></li>
		<li>Point three
	<ul>
		<li>Sub point 1</li>
		<li>Sub point 2</li>
	</ul></li>
	</ul>`,

	`@array[4] = 8@`,
	`	<p><code>array[4] = 8</code></p>`,

	`#{color:blue} one
# two
# three`,
	`	<ol style="color:blue;">
		<li>one</li>
		<li>two</li>
		<li>three</li>
	</ol>`,

	`Links (like "this":http://foo.com), are now mangled in 2.1.0, whereas 2.0 parsed them correctly.`,
	`	<p>Links (like <a href="http://foo.com">this</a>), are now mangled in 2.1.0, whereas 2.0 parsed them correctly.</p>`,

	`@monospaced text@, followed by text`,
	`	<p><code>monospaced text</code>, followed by text</p>`,

	`h2. A header





some text`,
	`	<h2>A header</h2>

	<p>some text</p>`,

	`*:(foo)foo bar baz*`,
	`	<p><strong cite="foo">foo bar baz</strong></p>`,

	`pre.. foo bar baz
quux`,
	`<pre>foo bar baz
quux
</pre>`,

	`line of text

    leading spaces`,
	`	<p>line of text</p>

    leading spaces`,

	`"some text":http://www.example.com/?q=foo%20bar and more text`,
	`	<p><a href="http://www.example.com/?q=foo%20bar">some text</a> and more text</p>`,

	`(??some text??)`,
	`	<p>(<cite>some text</cite>)</p>`,

	`(*bold text*)`,
	`	<p>(<strong>bold text</strong>)</p>`,

	`H[~2~]O`,
	`	<p>H<sub>2</sub>O</p>`,

	`p=. Où est l'école, l'église s'il vous plaît?`,
	`	<p style="text-align:center;">Où est l&#8217;école, l&#8217;église s&#8217;il vous plaît?</p>`,

	`p=. *_The_* _*Prisoner*_`,
	`	<p style="text-align:center;"><strong><em>The</em></strong> <em><strong>Prisoner</strong></em></p>`,

	`p=. "An emphasised _word._" & "*A spanned phrase.*" `,
	`	<p style="text-align:center;">&#8220;An emphasised <em>word.</em>&#8221; &amp; &#8220;<strong>A spanned phrase.</strong>&#8221; </p>`,

	`p=. "*Here*'s a word!" `,
	`	<p style="text-align:center;">&#8220;<strong>Here</strong>&#8217;s a word!&#8221; </p>`,

	`p=. "Please visit our "Textile Test Page":http://textile.sitemonks.com" `,
	`	<p style="text-align:center;">&#8220;Please visit our <a href="http://textile.sitemonks.com">Textile Test Page</a>&#8221; </p>`,

	`p=. Tell me, what is AJAX(Asynchronous Javascript and XML), please?`,
	`	<p style="text-align:center;">Tell me, what is <acronym title="Asynchronous Javascript and XML"><span class="caps">AJAX</span></acronym>, please?</p>`,

	`p{font-size:0.8em}. *TxStyle* is a documentation project of Textile 2.4 for "Textpattern CMS":http://texpattern.com.`,
	`	<p style="font-size:0.8em;"><strong>TxStyle</strong> is a documentation project of Textile 2.4 for <a href="http://texpattern.com">Textpattern <span class="caps">CMS</span></a>.</p>`,

	`"Übermensch":http://de/wikipedia.org/wiki/Übermensch`,
	`	<p><a href="http://de/wikipedia.org/wiki/%C3%9Cbermensch">Übermensch</a></p>`,

	`Here is some text with a <!-- Commented out[1] --> block.

<!-- Here is a single <span>line</span> comment block -->

<!-- Here is a whole
multiline
<span>HTML</span>
Comment
-->

bc. <!-- Here is a comment block in a code block. -->`,
	`	<p>Here is some text with a<!-- Commented out[1] --> block.</p>

	<p><!-- Here is a single <span>line</span> comment block --></p>

	<p><!-- Here is a whole
multiline
<span>HTML</span>
Comment
--></p>

<pre><code>&lt;!-- Here is a comment block in a code block. --&gt;
</code></pre>`,

	`"Textile(c)" is a registered(r) 'trademark' of Textpattern(tm) -- or TXP(That's textpattern!) -- at least it was - back in '88 when 2x4 was (+/-)5(o)C ... QED!

p{font-size: 200%;}. 2(1/4) 3(1/2) 4(3/4)`,
	`	<p>&#8220;Textile&#169;&#8221; is a registered&#174; &#8216;trademark&#8217; of Textpattern&#8482; &#8212; or <acronym title="That&#8217;s textpattern!"><span class="caps">TXP</span></acronym> &#8212; at least it was &#8211; back in &#8217;88 when 2&#215;4 was &#177;5&#176;C &#8230; <span class="caps">QED</span>!</p>

	<p style="font-size: 200%;">2&#188; 3&#189; 4&#190;</p>`,

	`|=. Testing colgroup and col syntax
|:\5. 80
|a|b|c|d|e|

|=. Testing colgroup and col syntax|
|:\5. 80|
|a|b|c|d|e|`,
	`	<table>
	<caption>Testing colgroup and col syntax</caption>
	<colgroup span="5" width="80">
	</colgroup>
		<tr>
			<td>a</td>
			<td>b</td>
			<td>c</td>
			<td>d</td>
			<td>e</td>
		</tr>
	</table>

	<table>
	<caption>Testing colgroup and col syntax</caption>
	<colgroup span="5" width="80">
	</colgroup>
		<tr>
			<td>a</td>
			<td>b</td>
			<td>c</td>
			<td>d</td>
			<td>e</td>
		</tr>
	</table>`,

	`table(#dvds){border-collapse:collapse}. Great films on DVD employing Textile summary, caption, thead, tfoot, two tbody elements and colgroups
|={font-size:140%;margin-bottom:15px}. DVDs with two Textiled tbody elements
|:\3. 100 |{background:#ddd}|250||50|300|
|^(header).
|_. Title |_. Starring |_. Director |_. Writer |_. Notes |
|~(footer).
|\5=. This is the tfoot, centred |
|-(toplist){background:#c5f7f6}.
| _The Usual Suspects_ | Benicio Del Toro, Gabriel Byrne, Stephen Baldwin, Kevin Spacey | Bryan Singer | Chris McQaurrie | One of the finest films ever made |
| _Se7en_ | Morgan Freeman, Brad Pitt, Kevin Spacey | David Fincher | Andrew Kevin Walker | Great psychological thriller |
| _Primer_ | David Sullivan, Shane Carruth | Shane Carruth | Shane Carruth | Amazing insight into trust and human psychology <br />rather than science fiction. Terrific! |
| _District 9_ | Sharlto Copley, Jason Cope | Neill Blomkamp | Neill Blomkamp, Terri Tatchell | Social commentary layered on thick,
but boy is it done well |
|-(medlist){background:#e7e895;}.
| _Arlington Road_ | Tim Robbins, Jeff Bridges | Mark Pellington | Ehren Kruger | Awesome study in neighbourly relations |
| _Phone Booth_ | Colin Farrell, Kiefer Sutherland, Forest Whitaker | Joel Schumacher | Larry Cohen | Edge-of-the-seat stuff in this
short but brilliantly executed thriller |`,
	`	<table style="border-collapse:collapse;" id="dvds" summary="Great films on DVD employing Textile summary, caption, thead, tfoot, two tbody elements and colgroups">
	<caption style="font-size:140%; margin-bottom:15px;"><span class="caps">DVD</span>s with two Textiled tbody elements</caption>
	<colgroup span="3" width="100">
	<col style="background:#ddd;" />
	<col width="250" />
	<col />
	<col width="50" />
	<col width="300" />
	</colgroup>
	<thead class="header">
		<tr>
			<th>Title </th>
			<th>Starring </th>
			<th>Director </th>
			<th>Writer </th>
			<th>Notes </th>
		</tr>
	</thead>
	<tfoot class="footer">
		<tr>
			<td style="text-align:center;" colspan="5">This is the tfoot, centred </td>
		</tr>
	</tfoot>
	<tbody style="background:#c5f7f6;" class="toplist">
		<tr>
			<td> <em>The Usual Suspects</em> </td>
			<td> Benicio Del Toro, Gabriel Byrne, Stephen Baldwin, Kevin Spacey </td>
			<td> Bryan Singer </td>
			<td> Chris McQaurrie </td>
			<td> One of the finest films ever made </td>
		</tr>
		<tr>
			<td> <em>Se7en</em> </td>
			<td> Morgan Freeman, Brad Pitt, Kevin Spacey </td>
			<td> David Fincher </td>
			<td> Andrew Kevin Walker </td>
			<td> Great psychological thriller </td>
		</tr>
		<tr>
			<td> <em>Primer</em> </td>
			<td> David Sullivan, Shane Carruth </td>
			<td> Shane Carruth </td>
			<td> Shane Carruth </td>
			<td> Amazing insight into trust and human psychology <br />
rather than science fiction. Terrific! </td>
		</tr>
		<tr>
			<td> <em>District 9</em> </td>
			<td> Sharlto Copley, Jason Cope </td>
			<td> Neill Blomkamp </td>
			<td> Neill Blomkamp, Terri Tatchell </td>
			<td> Social commentary layered on thick,<br />
but boy is it done well </td>
		</tr>
	</tbody>
	<tbody style="background:#e7e895;" class="medlist">
		<tr>
			<td> <em>Arlington Road</em> </td>
			<td> Tim Robbins, Jeff Bridges </td>
			<td> Mark Pellington </td>
			<td> Ehren Kruger </td>
			<td> Awesome study in neighbourly relations </td>
		</tr>
		<tr>
			<td> <em>Phone Booth</em> </td>
			<td> Colin Farrell, Kiefer Sutherland, Forest Whitaker </td>
			<td> Joel Schumacher </td>
			<td> Larry Cohen </td>
			<td> Edge-of-the-seat stuff in this<br />
short but brilliantly executed thriller </td>
		</tr>
	</tbody>
	</table>`,

	`-(hot) *coffee* := Hot _and_ black
-(hot#tea) tea := Also hot, but a little less black
-(cold) milk := Nourishing beverage for baby cows.
Cold drink that goes great with cookies. =:

-(hot) coffee := Hot and black
-(hot#tea) tea := Also hot, but a little less black
-(cold) milk :=
Nourishing beverage for baby cows.
Cold drink that goes great with cookies. =:`,
	`<dl>
	<dt class="hot"><strong>coffee</strong></dt>
	<dd>Hot <em>and</em> black</dd>
	<dt class="hot" id="tea">tea</dt>
	<dd>Also hot, but a little less black</dd>
	<dt class="cold">milk</dt>
	<dd>Nourishing beverage for baby cows.<br />
Cold drink that goes great with cookies.</dd>
</dl>

<dl>
	<dt class="hot">coffee</dt>
	<dd>Hot and black</dd>
	<dt class="hot" id="tea">tea</dt>
	<dd>Also hot, but a little less black</dd>
	<dt class="cold">milk</dt>
	<dd><p>Nourishing beverage for baby cows.<br />
Cold drink that goes great with cookies.</p></dd>
</dl>`,

	`;(class#id) Term 1
: Def 1
: Def 2
: Def 3`,
	`	<dl class="class" id="id">
		<dt>Term 1</dt>
		<dd>Def 1</dd>
		<dd>Def 2</dd>
		<dd>Def 3</dd>
	</dl>`,

	`*Here is a comment*

Here is *(class)a comment*

*(class)Here is a class* that is a little extended and is
*followed* by a strong word!

bc. ; Content-type: text/javascript
; Cache-Control: no-store, no-cache, must-revalidate, pre-check=0, post-check=0, max-age=0
; Expires: Sat, 24 Jul 2003 05:00:00 GMT
; Last-Modified: Wed, 1 Jan 2025 05:00:00 GMT
; Pragma: no-cache

*123 test*

*test 123*

**123 test**

**test 123**`,
	`	<p><strong>Here is a comment</strong></p>

	<p>Here is <strong class="class">a comment</strong></p>

	<p><strong class="class">Here is a class</strong> that is a little extended and is<br />
<strong>followed</strong> by a strong word!</p>

<pre><code>; Content-type: text/javascript
; Cache-Control: no-store, no-cache, must-revalidate, pre-check=0, post-check=0, max-age=0
; Expires: Sat, 24 Jul 2003 05:00:00 GMT
; Last-Modified: Wed, 1 Jan 2025 05:00:00 GMT
; Pragma: no-cache
</code></pre>

	<p><strong>123 test</strong></p>

	<p><strong>test 123</strong></p>

	<p><b>123 test</b></p>

	<p><b>test 123</b></p>`,

	`#_(first#list) one
# two
# three

test

#(ordered#list2).
# one
# two
# three

test

#_(class_4).
# four
# five
# six

test

#_ seven
# eight
# nine

test

# one
# two
# three

test

#22 22
# 23
# 24`,
	`	<ol class="first" id="list" start="1">
		<li>one</li>
		<li>two</li>
		<li>three</li>
	</ol>

	<p>test</p>

	<ol class="ordered" id="list2">
		<li>one</li>
		<li>two</li>
		<li>three</li>
	</ol>

	<p>test</p>

	<ol class="class_4" start="4">
		<li>four</li>
		<li>five</li>
		<li>six</li>
	</ol>

	<p>test</p>

	<ol start="7">
		<li>seven</li>
		<li>eight</li>
		<li>nine</li>
	</ol>

	<p>test</p>

	<ol>
		<li>one</li>
		<li>two</li>
		<li>three</li>
	</ol>

	<p>test</p>

	<ol start="22">
		<li>22</li>
		<li>23</li>
		<li>24</li>
	</ol>`,

	`# one
##3 one.three
## one.four
## one.five
# two

test

#_(continuation#section2).
# three
# four
##_ four.six
## four.seven
# five

test

#21 twenty-one
# twenty-two`,
	`	<ol>
		<li>one
	<ol start="3">
		<li>one.three</li>
		<li>one.four</li>
		<li>one.five</li>
	</ol></li>
		<li>two</li>
	</ol>

	<p>test</p>

	<ol class="continuation" id="section2" start="3">
		<li>three</li>
		<li>four
	<ol start="6">
		<li>four.six</li>
		<li>four.seven</li>
	</ol></li>
		<li>five</li>
	</ol>

	<p>test</p>

	<ol start="21">
		<li>twenty-one</li>
		<li>twenty-two</li>
	</ol>`,

	`|* Foo[^2^]
* _bar_
* ~baz~ |
|#4 *Four*
# __Five__ |
|-(hot) coffee := Hot and black
-(hot#tea) tea := Also hot, but a little less black
-(cold) milk :=
Nourishing beverage for baby cows.
Cold drink that goes great with cookies. =:
|`,
	`	<table>
		<tr>
			<td>	<ul>
		<li>Foo<sup>2</sup></li>
		<li><em>bar</em></li>
		<li><sub>baz</sub></li>
	</ul></td>
		</tr>
		<tr>
			<td>	<ol start="4">
		<li><strong>Four</strong></li>
		<li><i>Five</i></li>
	</ol></td>
		</tr>
		<tr>
			<td><dl>
	<dt class="hot">coffee</dt>
	<dd>Hot and black</dd>
	<dt class="hot" id="tea">tea</dt>
	<dd>Also hot, but a little less black</dd>
	<dt class="cold">milk</dt>
	<dd><p>Nourishing beverage for baby cows.<br />
Cold drink that goes great with cookies.</p></dd><br />
</dl></td>
		</tr>
	</table>`,

	`h4. A more complicated table

table(tableclass#tableid){color:blue}.
|_. table |_. more |_. badass |
|\3. Horizontal span of 3|
(firstrow). |first|HAL(open the pod bay doors)|1|
|some|{color:green}. styled|content|
|/2. spans 2 rows|this is|quite a|
| deep test | don't you think?|
(lastrow). |fifth|I'm a lumberjack|5|
|sixth| _*bold italics*_ |6|`,
	`	<h4>A more complicated table</h4>

	<table style="color:blue;" class="tableclass" id="tableid">
		<tr>
			<th>table </th>
			<th>more </th>
			<th>badass </th>
		</tr>
		<tr>
			<td colspan="3">Horizontal span of 3</td>
		</tr>
		<tr class="firstrow">
			<td>first</td>
			<td><acronym title="open the pod bay doors"><span class="caps">HAL</span></acronym></td>
			<td>1</td>
		</tr>
		<tr>
			<td>some</td>
			<td style="color:green;">styled</td>
			<td>content</td>
		</tr>
		<tr>
			<td rowspan="2">spans 2 rows</td>
			<td>this is</td>
			<td>quite a</td>
		</tr>
		<tr>
			<td> deep test </td>
			<td> don&#8217;t you think?</td>
		</tr>
		<tr class="lastrow">
			<td>fifth</td>
			<td>I&#8217;m a lumberjack</td>
			<td>5</td>
		</tr>
		<tr>
			<td>sixth</td>
			<td> <em><strong>bold italics</strong></em> </td>
			<td>6</td>
		</tr>
	</table>`,

	`| *strong* |

| _em_ |

| Inter-word -dashes- | ZIP-codes are 5- or 9-digit codes |`,
	`	<table>
		<tr>
			<td> <strong>strong</strong> </td>
		</tr>
	</table>

	<table>
		<tr>
			<td> <em>em</em> </td>
		</tr>
	</table>

	<table>
		<tr>
			<td> Inter-word <del>dashes</del> </td>
			<td> <span class="caps">ZIP</span>-codes are 5- or 9-digit codes </td>
		</tr>
	</table>`,

	`|_. attribute list |
|<. align left |
|>. align right|
|=. center |
|<>. justify me|
|^. valign top |
|~. bottom |`,
	`	<table>
		<tr>
			<th>attribute list </th>
		</tr>
		<tr>
			<td style="text-align:left;">align left </td>
		</tr>
		<tr>
			<td style="text-align:right;">align right</td>
		</tr>
		<tr>
			<td style="text-align:center;">center </td>
		</tr>
		<tr>
			<td style="text-align:justify;">justify me</td>
		</tr>
		<tr>
			<td style="vertical-align:top;">valign top </td>
		</tr>
		<tr>
			<td style="vertical-align:bottom;">bottom </td>
		</tr>
	</table>`,

	`h2. A definition list

;(class#id) Term 1
: Def 1
: Def 2
: Def 3
;; Center
;; NATO(Why Em Cee Ayy)
:: Subdef 1
:: Subdef 2
;;; SubSub Term
::: SubSub Def 1
::: SubSub Def 2
::: Subsub Def 3
With newline
::: Subsub Def 4
:: Subdef 3
: DEF 4
; Term 2
: Another def
: And another
: One more
:: A def without a term
:: More defness
; Third term for good measure
: My definition of a boombastic jazz`,
	`	<h2>A definition list</h2>

	<dl class="class" id="id">
		<dt>Term 1</dt>
		<dd>Def 1</dd>
		<dd>Def 2</dd>
		<dd>Def 3
	<dl>
		<dt>Center</dt>
		<dt><acronym title="Why Em Cee Ayy"><span class="caps">NATO</span></acronym></dt>
		<dd>Subdef 1</dd>
		<dd>Subdef 2
	<dl>
		<dt>SubSub Term</dt>
		<dd>SubSub Def 1</dd>
		<dd>SubSub Def 2</dd>
		<dd>Subsub Def 3<br />
With newline</dd>
		<dd>Subsub Def 4</dd>
	</dl></dd>
		<dd>Subdef 3</dd>
	</dl></dd>
		<dd><span class="caps">DEF</span> 4</dd>
		<dt>Term 2</dt>
		<dd>Another def</dd>
		<dd>And another</dd>
		<dd>One more
	<dl>
		<dd>A def without a term</dd>
		<dd>More defness</dd>
	</dl></dd>
		<dt>Third term for good measure</dt>
		<dd>My definition of a boombastic jazz</dd>
	</dl>`,

	`###. Here's a comment.

h3. Hello

###. And
another
one.

Goodbye.`,
	`	<h3>Hello</h3>

	<p>Goodbye.</p>`,

	`h2. A Definition list which covers the instance where a new definition list is created with a term without a definition

- term :=
- term2 := def`,
	`	<h2>A Definition list which covers the instance where a new definition list is created with a term without a definition</h2>

<dl>
	<dt>term2</dt>
	<dd>def</dd>
</dl>`,
}
