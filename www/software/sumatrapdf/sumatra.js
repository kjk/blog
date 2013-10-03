// update after releasing a new version
var gSumVer = "2.4";

// used by download-prev* pages, update after releasing a new version
var gPrevSumatraVersion = [
	"2.3.3", "2.3.1", "2.3", "2.2.1", "2.2", "2.1.1",
	"2.1", "2.0.1", "2.0", "1.9", "1.8", "1.7", "1.6",
	"1.5.1", "1.5", "1.4", "1.3", "1.2", "1.1", "1.0.1",
	"1.0", "0.9.4", "0.9.3", "0.9.1", "0.9", "0.8.1",
	"0.8", "0.7", "0.6", "0.5", "0.4", "0.3", "0.2"
];

var gSumZipUrl = "https://kjkpub.s3.amazonaws.com/sumatrapdf/rel/SumatraPDF-" + gSumVer + ".zip";
var gSumExeUrl = "https://kjkpub.s3.amazonaws.com/sumatrapdf/rel/SumatraPDF-" + gSumVer + "-install.exe";
var gSumZipName = "SumatraPDF-" + gSumVer + ".zip";
var gSumExeName = "SumatraPDF-" + gSumVer + "-install.exe";

// used by download-free-pdf-viewer*.html pages
function dlHtml(s1,s2, s3) {
	if (!s3) {
		s3 = "";
	} else {
		s3 = " <span style='font-size:90%; color:gray'>" + s3 + "</span>";
	}
	return '<table><tr><td>' + s1 + '&nbsp;&nbsp;</td><td><a href="' +
	gSumExeUrl + '" onclick="return SetupRedirect()">' + gSumExeName +
	'</a></td></tr><tr><td>' + s2 + '&nbsp;&nbsp;</td><td><a href="' +
	gSumZipUrl + '" onclick="return SetupRedirect()">' + gSumZipName +
	'</a>' + s3 + '</td></tr></table>';
}

// used by downloadafter*.html pages
function dlAfterHtml(s1,s2,s3,s4) {
    return '<a href="' + gSumExeUrl + '">' + s1 + '</a>' + s2 +
    '<a href="' + gSumZipUrl + '">' + s3 + '</a>' + s4;
}

var gRuTrans = {
	"Home" : "Начало",
	"Version History" : "Новости",
	"Manual" : "Руководство пользователя",
	"Download" : "Загрузка",
	"Contribute" : "Сотрудничество",
	"Translations" : "Переводы",
	"Forums" : "Форум"
};

var gBgTrans = {
	"Home" : "Начало",
	"Version History" : "Новини",
	"Manual" : "Ръководство",
	"Download" : "Сваляне",
	"Contribute" : "Допринесете",
	"Translations" : "Преводи",
	"Forums" : "Форум"
};

var gRoTrans = {
	"Home" : "Acasă",
	"Version History" : "Ştiri",
	"Manual" : "Manual",
	"Download" : "Descarcă",
	"Contribute" : "Contribuie",
	"Translations" : "Traduceri",
	"Forums" : "Forum"
};

var gPtTrans = {
	"Home" : "Página inicial",
	"Version History" : "Novidades",
	"Manual" : "Manual",
	"Download" : "Transferências",
	"Contribute" : "Participar",
	"Translations" : "Traduções",
	"Forums" : "Fóruns"
};

var gJaTrans = {
	"Home" : "ホーム",
	"Version History" : "ニュース",
	"Manual" : "マニュアル",
	"Download" : "ダウンロード",
	"Contribute" : "開発",
	"Translations" : "翻訳",
	"Forums" : "フォーラム"
};

var gEsTrans = {
	"Home" : "Home",
	"Version History" : "Noticias",
	"Manual" : "Manual",
	"Download" : "Descargar",
	"Contribute" : "Contribuir",
	"Translations" : "Traducciones",
	"Forums" : "Foros"
};

var gDeTrans = {
	"Home" : "Home",
	"Version History" : "Neuigkeiten",
	"Manual" : "Handbuch",
	"Download" : "Download",
	"Contribute" : "Helfen",
	"Translations" : "Übersetzungen",
	"Forums" : "Forum"
};

var gFrTrans = {
	"Home" : "Accueil",
	"Version History" : "Historique",
	"Manual" : "Manuel",
	"Download" : "T&eacute;l&eacute;charger",
	"Contribute" : "Contribuer",
	"Translations" : "Traductions",
	"Forums" : "Forum"
};

var gCnTrans = {
	"Home" : "主页",
	"Version History" : "新闻",
	"Manual" : "手册",
	"Download" : "下载",
	"Contribute" : "参与贡献",
	"Translations" : "翻译",
	"Forums" : "论坛"
};

var gSrTrans = {
	"Home" : "Почетак",
	"Version History" : "Новости",
	"Manual" : "Упутство",
	"Download" : "Преузимање",
	"Contribute" : "Допринос",
	"Translations" : "Преводи",
	"Forums" : "Форум"
};

var gKaTrans = {
	"Home" : "მთავარი",
	"Version History" : "სიახლეები",
	"Manual" : "სახელმძღვანელო",
	"Download" : "ჩამოტვირთვა",
	"Contribute" : "მონაწილეობა",
	"Translations" : "თარგმნები",
	"Forums" : "ფორუმები"
};

var gEuTrans = {
	"Home" : "Hasiera",
	"Version History" : "Bertsio Historia",
	"Manual" : "Eskuliburua",
	"Download" : "Jeisketa",
	"Contribute" : "Ekarpenak",
	"Translations" : "Itzulpenak",
	"Forums" : "Foroak"
};

var gUzTrans = {
	"Home" : "Bosh sahifa",
	"Version History" : "Versiyalar tarixi",
	"Manual" : "Qo'llanma",
	"Download" : "Yuklab olish",
	"Contribute" : "Ishtirok etish",
	"Translations" : "Tarjimalar",
	"Forums" : "Forum"
};

var gHrTrans = {
	"Home" : "Početna",
	"Version History" : "Povijest verzija",
	"Manual" : "Priručnik",
	"Download" : "Preuzimanje",
	"Contribute" : "Sudjelovanje",
	"Translations" : "Prijevod",
	"Forums" : "Forum"
};

var gTabTrans = {
	"ru" : gRuTrans, "ro" : gRoTrans, "pt" : gPtTrans,
	"ja" : gJaTrans, "es" : gEsTrans, "de" : gDeTrans,
	"fr" : gFrTrans, "cn" : gCnTrans, "bg" : gBgTrans,
	"sr" : gSrTrans, "ka" : gKaTrans, "eu" : gEuTrans,
	"uz" : gUzTrans, "hr" : gHrTrans
};

var gLangCookieName = "forceSumLang";

// we also use the order of languages in this array
// to order links for translated pages
var gLanguages = [
	"en", ["English", "English"],
	"bg", ["Български", "Bulgarian"],
	"sr", ["Српски", "Serbian"],
	"de", ["Deutsch", "German"],
	"es", ["Español", "Spanish"],
	"eu", ["Euskara", "Basque"],
	"fr", ["Français", "French"],
	"hr", ["Hrvatski", "Croatian"],
	"pt", ["Português", "Portuguese"],
	"ru", ["Pусский", "Russian"],
	"ro", ["Română", "Romanian"],
	"ja", ["日本語", "Japanese"],
	"cn", ["简体中文", "Chinese"],
	"ka", ["ქართული ენა", "Georgian"],
	"uz", ["O'zbek", "Uzbek"]
];

// For ie compat, from https://developer.mozilla.org/en/JavaScript/Reference/Global_Objects/Array/indexOf
if (!Array.prototype.indexOf) {
    Array.prototype.indexOf = function (searchElement) {
        "use strict";
        if (this === void 0 || this === null) {
            throw new TypeError();
        }
        var t = Object(this);
        var len = t.length >>> 0;
        if (len === 0) {
            return -1;
        }
        var n = 0;
        if (arguments.length > 0) {
            n = Number(arguments[1]);
            if (n !== n) { // shortcut for verifying if it's NaN
                n = 0;
            } else if (n !== 0 && n !== Infinity && n !== -Infinity) {
                n = (n > 0 || -1) * Math.floor(Math.abs(n));
            }
        }
        if (n >= len) {
            return -1;
        }
        var k = n >= 0 ? n : Math.max(len - Math.abs(n), 0);
        for (; k < len; k++) {
            if (k in t && t[k] === searchElement) {
                return k;
            }
        }
        return -1;
    }
}

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
	alert("No native name for lang '" + lang + "'");
	return "";
}

var gTransalatedPages = [
	"download-free-pdf-viewer", ["ka", "ru", "cn", "de", "es", "fr", "ja", "pt", "ro", "ru", "bg", "sr", "eu", "uz", "hr"],
	"download-prev", ["ka", "de", "fr", "es", "ja", "pt", "ro", "sr", "eu", "uz", "hr"],
	"downloadafter", ["ka", "de", "fr", "es", "ja", "pt", "ro", "bg", "sr", "eu", "uz", "hr"],
	"free-pdf-reader", ["ka", "cn", "de", "fr", "es", "ja", "pt", "ro", "ru", "bg", "sr", "eu", "uz", "hr"],
	"manual", ["ka", "ru", "cn", "de", "fr", "es", "ja", "pt", "ro", "ru", "bg", "sr", "eu", "uz", "hr"]
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

// A heuristic used to detect preferred language of the user
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

function autoRedirectToTranslated() {
	var tmp = getBaseUrlAndLang();
	var baseUrl = tmp[0];
	var pageLang = tmp[1];
	// only redirect if we're on an english page
	if (!isEng(pageLang)) {
		alert("autoRedirectToTranslated() called from non-english page "+tmp);
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
			<li><a href="http://code.google.com/p/sumatrapdf/wiki/JoinTheProject" title="Contribute"><span>Сотрудничество</span></a></li>
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
		["download-free-pdf-viewer", "Download"],
		["manual", "Manual"],
		["news", "Version History"],
		//["develop", "Contribute"],
		//["translations", "Translations"],
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
	return '';
/*
return '<span style="position:relative; left: 22px; top: 6px;">\
<script type="text/javascript" src="http://apis.google.com/js/plusone.js"></script>\
<g:plusone size="medium" href="http://blog.kowalczyk.info/software/sumatrapdf/"></g:plusone>\
</span>';
*/
}

function buttonsOldHtml() {
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

/* adzerk ads */

var OS_UNKNOWN = -1;
var OS_WIN_XP = 0;
var OS_WIN_VISTA = 1;
var OS_WIN_7 = 2;
var OS_WIN_8 = 3;
function getOS() {
	var userAgent = navigator.userAgent;
	if (-1 != userAgent.indexOf("Windows NT 5.1")) {
		return OS_WIN_XP;
	}
	if (-1 != userAgent.indexOf("Windows NT 6.0")) {
		return OS_WIN_VISTA;
	}
	if (-1 != userAgent.indexOf("Windows NT 6.1")) {
		return OS_WIN_7;
	}
	if (-1 != userAgent.indexOf("Windows NT 6.2")) {
		return OS_WIN_8;
	}
	return OS_UNKNOWN;
};

// slimware-01
// http://kkowalczyk.adzerk.com/network/7803/brand/33242/campaign/48296/option/86975/creatives/129416/map/182468 xp
// http://kkowalczyk.adzerk.com/network/7803/brand/33242/campaign/48296/option/86975/creatives/129418/map/182469 vista
// http://kkowalczyk.adzerk.com/network/7803/brand/33242/campaign/48296/option/86975/creatives/129420/map/182470 win7
// http://kkowalczyk.adzerk.com/network/7803/brand/33242/campaign/48296/option/86975/creatives/129421/map/182471 win8
function getCreativeId() {
	var os = getOS();
	if (OS_WIN_XP == os) {
		return 182468;
	} else if (OS_WIN_VISTA == os) {
		return 182469;
	} else if (OS_WIN_7 == os) {
		return 182470;
	} else if (OS_WIN_8 == os) {
		return 182471;
	}
	return 0;
}

var creativeByOs = [182468, 182469, 182470, 182471];
function getCreativeId2() {
	var os = getOS();
	if (OS_UNKNOWN != os) {
		return creativeByOs(os);
	}
	return 0;
}

var ados = ados || {};
function doAdzerk() {
	ados.run = ados.run || [];
	var creativeId = getCreativeId();
	//creativeId = 182470;
	console.log("creativeId = " + creativeId);
	if (0 == creativeId) {
		doAdsense();
		return;
	}

	ados.run.push(function() {
		/* load placement for account: kkowalczyk, site: blog, size: 728x90 - Leaderboard*/
		ados_add_placement(7803, 51221, "azk9318", 4).setFlightCreativeId(creativeId);
		ados_load();
	});
}

function doAdzerk2() {
	ados.run = ados.run || [];
	ados.run.push(function() {
		/* load placement for account: kkowalczyk, site: blog, size: 728x90 - Leaderboard*/
		ados_add_placement(7803, 51221, "azk9318", 4);
		ados_load();
	});
}

function doAdsense() {
	(adsbygoogle = window.adsbygoogle || []).push({});
}

function dispatchAd() {
	var os = getOS();
	if (-1 == os) {
		doAdsense();
	} else {
		doAdzerk();
	}
}
