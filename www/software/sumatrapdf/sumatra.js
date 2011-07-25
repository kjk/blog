var gRuTrans = {
	"Home" : "Начало",
	"News" : "Новости",
	"Manual" : "Руководство пользователя",
	"Download" : "Загрузка",
	"Contribute" : "Сотрудничество",
	"Translations" : "Переводы",
	"Forums" : "Форум"
};

var gBgTrans = {
	"Home" : "Начало",
	"News" : "Новини",
	"Manual" : "Ръководство",
	"Download" : "Сваляне",
	"Contribute" : "Допринесете",
	"Translations" : "Преводи",
	"Forums" : "Форум"
};

var gRoTrans = {
	"Home" : "Acasă",
	"News" : "Ştiri",
	"Manual" : "Manual",
	"Download" : "Descarcă",
	"Contribute" : "Contribuie",
	"Translations" : "Traduceri",
	"Forums" : "Forum"
};

var gPtTrans = {
	"Home" : "Página inicial",
	"News" : "Novidades",
	"Manual" : "Manual",
	"Download" : "Transferências",
	"Contribute" : "Participar",
	"Translations" : "Traduções",
	"Forums" : "Fóruns"
};

var gJaTrans = {
	"Home" : "ホーム",
	"News" : "ニュース",
	"Manual" : "マニュアル",
	"Download" : "ダウンロード",
	"Contribute" : "開発",
	"Translations" : "翻訳",
	"Forums" : "フォーラム"
};

var gEsTrans = {
	"Home" : "Home",
	"News" : "Noticias",
	"Manual" : "Manual",
	"Download" : "Descargar",
	"Contribute" : "Contribuir",
	"Translations" : "Traducciones",
	"Forums" : "Foros"
};

var gDeTrans = {
	"Home" : "Home",
	"News" : "Neuigkeiten",
	"Manual" : "Handbuch",
	"Download" : "Download",
	"Contribute" : "Helfen",
	"Translations" : "Übersetzungen",
	"Forums" : "Forum"	
};

var gCnTrans = {
	"Home" : "主页",
	"News" : "新闻",
	"Manual" : "手册",
	"Download" : "下载",
	"Contribute" : "参与贡献",
	"Translations" : "翻译",
	"Forums" : "论坛"	
};

var gTabTrans = {
	"ru" : gRuTrans, "ro" : gRoTrans, "pt" : gPtTrans, "ja" : gJaTrans,
	"es" : gEsTrans, "de" : gDeTrans, "cn" : gCnTrans, "bg" : gBgTrans
};

var gLangCookieName = "forceSumLang";

// we also use the order of languages in this array
// to order links for translated pages
var gLanguages = [
	"en", ["English", "English"],
	"de", ["Deutsch", "German"],
	"es", ["Español", "Spanish"],
	//"fr", ["Français", "French"],
	"pt", ["Português", "Portuguese"],
	"ru", ["Pусский", "Russian"],
	"bg", ["Български", "Bulgarian"],
	"ro", ["Română", "Romanian"],
	"ja", ["日本語", "Japanese"],
	"cn", ["简体中文", "Chinese"]
];

function isValidLang(lang) {
	return -1 != gLanguages.indexOf(lang);
}

function langNativeName(lang) {
	if ("" == lang) { return "English" };
	var i;
	for (i=0; i<gLanguages.length / 2; i++) {
		if (gLanguages[i*2] == lang) {
			return gLanguages[i*2+1][0];
		}
	}
	allert("No native name for lang '" + lang + "'");
	return "";
}

var gTransalatedPages = [
	"download-free-pdf-viewer", ["ru", "cn", "de", "es", "fr", "ja", "pt", "ro", "ru", "bg"],
	"download-prev", ["de", "es", "ja", "pt", "ro"],
	"downloadafter", ["de", "es", "ja", "pt", "ro", "bg"],
	"free-pdf-reader", ["cn", "de", "es", "ja", "pt", "ro", "ru", "bg"],
	"manual", ["ru", "cn", "de", "es", "ja", "pt", "ro", "ru", "bg"],
	"news", ["pt"]
];

// return a list of langauges that a given page is translated into
function translationsForPage(baseUrl) {
	var i;
	for (i=0; i<gTransalatedPages.length / 2; i++) {
		if (gTransalatedPages[i*2] == baseUrl) {
			return gTransalatedPages[i*2+1];
		}
	}
	return [];
}

function hasTranslation(baseUrl, lang) {
	return -1 != translationsForPage(baseUrl).indexOf(lang);
}

function setCookie(name,val,expireInDays) {
    var d=new Date();
    d.setDate(d.getDate()+expireInDays);
    document.cookie=name+"="+escape(val)+
    ((expireInDays==null) ? "" : ";expires="+d.toGMTString());
}

function getCookie(name) {
	var c = document.cookie;
	var start = c.indexOf(name + "=");
	if (-1 == start) { return null; }
    start += name.length+1;
    var end = c.indexOf(";", start);
    if (-1 == end) { end = c.length };
    return unescape(c.substring(start,end));
}

function deleteCookie(name) {
	setCookie(name, "", 0);
}

function langFromCookie() {
	var lang = getCookie(gLangCookieName);
	if (isValidLang(lang)) {
		return lang;
	}
	return null;
}

function deleteLangCookie() {
	deleteCookie(gLangCookieName);
}

function setLangCookie(lang) {
	setCookie(gLangCookieName, lang, 365);
}

// A heuristic used to detect preffered language of the user
// based on settings in navigator object.
// TODO: we should also look at Accept-Language header, which might look like:
// Accept-Language:ru,en-US;q=0.8,en;q=0.6
// but headers are not available from JavaScript so I would have to
// pass this data somehow from the server to html or use additional
// request from javascript as described at http://stackoverflow.com/questions/1043339/javascript-for-detecting-browser-language-preference
function detectUserLang() {
	var n = window.navigator;
	// a heuristic: userLanguage and browserLanguage are for ie
	// language is for FireFox and Chrome
	var lang1 = n.userLanguage || n.browserLanguage || n.language || "en";
	// we only care about "en" part of languages like "en-US"
	return lang1.substring(0,2);
}

function forceRedirectToLang(lang) {
	var tmp = getBaseUrlAndLang();
	var baseUrl = tmp[0];
	if (!hasTranslation(baseUrl, lang)) { lang = "en"; }
	window.location = urlFromBaseUrlLang(baseUrl, lang);
}

// TODO: should also redirect from non-english pages for consistency?
function autoRedirectToTranslated() {
	var tmp = getBaseUrlAndLang();
	var baseUrl = tmp[0];
	var pageLang = tmp[1];
	// only redirect if we
	if (!isEng(pageLang)) {
		alert("autoRedirectToTranslated() called from non-english page");
		return;
	}
	var cookieLang = langFromCookie();
	if (cookieLang) {
		if (cookieLang == pageLang) { return; }
		if (hasTranslation(baseUrl, cookieLang)) {
			window.location = urlFromBaseUrlLang(baseUrl, cookieLang);
		}
		return;
	}

	var userLang = detectUserLang();
	if (userLang == pageLang) { return; }
	if (!hasTranslation(baseUrl, userLang)) { return; }
	window.location = urlFromBaseUrlLang(baseUrl, userLang);
}

function cookieOrUserLang() {
	return langFromCookie() || detectUserLang();
}

// sumatra urls are in format:
// /software/sumatrapdf/${url}[-${lang}].html
// return ${url} and ${lang} parts
// ${lang} can be "" which means english (en)
function getBaseUrlAndLang() {
	var lang = "";
	var url = location.pathname.split("/");
	url = url[url.length-1];
	url = url.split(".html")[0];
	if (url[url.length-3] == '-') {
		lang = url.substring(url.length-2)
		url = url.substring(0, url.length-3);
	}
	//alert(url + "," + lang);
	return [url, lang];	
}

function isEng(lang) {
	return (lang == "") || (lang=="en");
}

// construct text like:
// <span class="trans"><a href="free-pdf-reader-de.html">Deutsch</a></span>
function langsLinkHtml(baseUrl, lang) {
	var url = baseUrl;
	if (!isEng(lang)) {
		url = url + "-" + lang;
	}
	url += ".html";
	return '<span class="trans"><a href="' + url + '">' + langNativeName(lang) + '</a></span>&nbsp;';
}

function sortByLang(l1, l2) {
	var l1Idx = gLanguages.indexOf(l1);
	var l2Idx = gLanguages.indexOf(l2);
	return l1Idx - l2Idx;
}

function langsNavHtmlOld() {
	var i, l, baseUrl, lang;
	var tmp = getBaseUrlAndLang();
	baseUrl = tmp[0];
	lang = tmp[1];
	var translations = translationsForPage(baseUrl);
	translations.sort(sortByLang);
	if (0 == translations.length) {
		// shouldn't happen becase should only be called from pages
		// that were translated
		alert("No translations for page " + baseUrl);
	}
	var s = '<span style="float: right;">';
	var l;
	if (!isEng(lang)) {
		s += langsLinkHtml(baseUrl, "en");
	}
	for (i=0; i<translations.length; i++) {
		l = translations[i];
		if (l == lang) {
			continue;
		}
		s += langsLinkHtml(baseUrl, l);
	}
	s += '</span>';
	return s;
}

// Generate this html:
/*
<span style="float: right; color: black; font-size: 80%;">
Language:
<select id=langSelect onchange="langChanged();">
  <option value="en">English</option>
  <option selected="selected" value="de">Deutsh</option>
  <option value="default">Default</option>
</select>
</span>
*/
function langsNavHtml() {
	var i, userLang, langName, issel;
	var userLang = cookieOrUserLang();
	var s = '<span style="float: right; color: black; font-size: 80%;">\
Language:\
<select id=langSelect onchange="langChanged();">'

	for (i=0; i<gLanguages.length / 2; i++) {
		issel = "";
		lang = gLanguages[i*2];
		if (userLang == lang) {
			issel = 'selected="selected" ';
		}
		langName = gLanguages[i*2+1][0];
		s += '<option ' + issel + 'value="' + lang + '">' + langName + '</option>';
	}
	s += '<option value="default">Default</option>';
	s += '</select></span>';
	return s;
};

function translateTabText(lang, s) {
	if (!gTabTrans[lang]) { return s; }
	return gTabTrans[lang][s] || s;
}

function urlFromBaseUrlLang(baseUrl, lang) {
	if (baseUrl == "/forum_sumatra/") {
		return baseUrl;
	}
	if (hasTranslation(baseUrl, lang)) {
		return baseUrl + "-" + lang + ".html";
	}
	return baseUrl + ".html";
}

/*
	Construct html as below, filling the apropriate inter-language links.
	<div id="ddcolortabs">
		<ul>
			<li id="current"><a href="free-pdf-reader.html" title="Home"><span>Начало</span></a></li>
			<li><a href="news.html" title="News"><span>Новости</span></a></li>
			<li><a href="manual.html" title="Manual"><span>Руководство пользователя</span></a></li>
			<li><a href="download-free-pdf-viewer.html" title="Download"><span>Загрузка</span></a></li>
			<li><a href="develop.html" title="Contribute"><span>Сотрудничество</span></a></li>
			<li><a href="translations.html" title="Translations"><span>Переводы</span></a></li>
			<li><a href="/forum_sumatra/" title="Forums"><span>Форум</span></a></li>
		</ul>
	</div>
	<div id="ddcolortabsline"> </div>
*/
function navHtml() {
	var i, baseUrl, lang, currUrl, txt, url;
	var tmp = getBaseUrlAndLang();
	baseUrl = tmp[0];
	lang = tmp[1];

	var s = '<div id="ddcolortabs"><ul>';
	var baseUrls = [
		["free-pdf-reader", "Home"],
		["news", "News"],
		["manual", "Manual"],
		["download-free-pdf-viewer", "Download"], 
		["develop", "Contribute"],
		["translations", "Translations"],
		["/forum_sumatra/", "Forums"]];

	for (i=0; i<baseUrls.length; i++) {
		currUrl = baseUrls[i][0];
		if (currUrl == baseUrl) {
			s += '<li id="current">';
		} else {
			s += '<li>';
		}
		txt = translateTabText(lang, baseUrls[i][1]);
		url = urlFromBaseUrlLang(currUrl, lang);
		s += '<a href="' + url + '" title="' + txt + '"><span>' + txt + '</span></a></li>';
	}
	
	s += '</ul></div><div id="ddcolortabsline"> </div>';
	return s;
}

function installerHref(ver) {
	return '<a href="http://kjkpub.s3.amazonaws.com/sumatrapdf/rel/SumatraPDF-' + ver + '-install.exe">SumatraPDF-' + ver + '-install.exe</a>';
}
function zipHref(ver) {
	return '<a href="http://kjkpub.s3.amazonaws.com/sumatrapdf/rel/SumatraPDF-' + ver + '.zip">SumatraPDF-' + ver + '.zip</a>';
}

// used by download-prev* pages
// Update after releasing a new version
var gPrevSumatraVersion = [
	"1.6", 
	"1.5.1", "1.5", "1.4", "1.3", "1.2", "1.1", "1.0.1",
	"1.0", "0.9.4", "0.9.3", "0.9.1", "0.9", "0.8.1", 
	"0.8", "0.7", "0.6", "0.5", "0.4", "0.3", "0.2"
];

// used by download-prev* pages
function prevLanguagesList(installerStr, zipFileStr) {
	var s, i;
	var s = "";
	for (i=0; i < gPrevSumatraVersion.length; i++)
	{
		var ver = gPrevSumatraVersion[i];
		s += '<p>' + installerStr + ': ' + installerHref(ver) + '<br>\n';
		s += zipFileStr + ': ' + zipHref(ver) + '</p>\n';
	}
	return s;        
}

function langChanged() {
	var dd = document.getElementById("langSelect");
    var idx  = dd.selectedIndex;
    var lang = dd.options[idx].value;
    //alert("Selected value: " + lang);
    if (lang == "default") {
    	deleteLangCookie();
	    forceRedirectToLang(detectUserLang());
    	return true;
    }
    setLangCookie(lang);
    forceRedirectToLang(lang);
    return true;
};

function buttonsHtml() {
return '<span style="position:relative; left: 22px; top: 6px;">\
<script type="text/javascript" src="http://apis.google.com/js/plusone.js"></script>\
<g:plusone size="medium" href="http://blog.kowalczyk.info/software/sumatrapdf/"></g:plusone>\
</span>\
<span style="position:relative; left: 12px; top: 6px;">\
<a href="http://twitter.com/share" class="twitter-share-button" data-url="http://blog.kowalczyk.info/software/sumatrapdf/free-pdf-reader.html" data-text="SumatraPDF - free PDF reader for Windows" data-count="horizontal" data-via="kjk">Tweet</a><script type="text/javascript" src="http://platform.twitter.com/widgets.js"></script>\
</span>\
<span style="position:relative; top: 7px; left: 0px;">\
<iframe src="http://www.facebook.com/plugins/like.php?href=http%3A%2F%2Fblog.kowalczyk.info%2Fsoftware%2Fsumatrapdf%2F&amp;layout=button_count&amp;show_faces=false&amp;width=450&amp;action=like&amp;colorscheme=light&amp;height=21" scrolling="no" frameborder="0" style="border:none; overflow:hidden; width:88px; height:21px;" allowTransparency="true"></iframe>\
</span>';
}

function yepiAdHtml() {
return '<center>\
<table class="ad" cellspacing=0 cellpadding=0>\
<tr><td>\
  <a href="http://www.yepi.ws/fotofi/free-stock-photos"><b>Free stock photos</b></a>\
</td></tr>\
<tr><td>\
<span style="font-size: 80%"><span class="adl">www.yepi.ws</span>&nbsp;&nbsp;&nbsp;&nbsp;\
Find free stock photos with Fotofi. More than 100 million photos available.</span>\
</td></tr>\
</table>\
</center>';
}

function googleAnalytics() {
    var _gaq = _gaq || [];
    _gaq.push(['_setAccount', 'UA-194516-1']);
    _gaq.push(['_trackPageview']);

    (function() {
      var ga = document.createElement('script'); ga.type = 'text/javascript'; ga.async = true;
      ga.src = ('https:' == document.location.protocol ? 'https://ssl' : 'http://www') + '.google-analytics.com/ga.js';
      (document.getElementsByTagName('head')[0] || document.getElementsByTagName('body')[0]).appendChild(ga);
    })();
}

function adData() {
	google_ad_client = "pub-8305375090017172";
	/* sumatra */
	google_ad_slot = "5091334157";
	google_ad_width = 728;
	google_ad_height = 90;
}
