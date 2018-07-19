package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kjk/u"
)

var articleRedirectsTxt = `3|article/Diet.html
36|article/objdump-g.html
141|article/gdb-quick-reference-1.html
178|article/Interface-Builder-reference.html
213|article/Mac-software-installed.html
217|article/backtrace_symbols-and-rdynamic-in-gcc.html
243|article/Mac-program-scheduling-like-crontab.html
263|article/Deliberate-practice.html
266|article/Exercise.html
296|article/Make-C-code-safe-for-C.html
296|kb/make-c-code-safe-for-c.html
301|article/Basics-of-writing-DOS-bat-batch-files.html
301|kb/basics-of-writing-dos-.bat-batch-files.html
302|article/Compile-time-asserts-in-C.html
302|kb/compile-time-asserts-in-c.html
303|article/Get-file-size-under-windows.html
303|kb/get-file-size-under-windows.html
304|article/Subversion-basics.html
304|kb/subversion-basics.html
305|article/Check-if-file-exists-on-Windows.html
305|kb/check-if-file-exists-on-windows.html
306|article/High-resolution-timer-for-timing-code-fragments.html
306|kb/high-resolution-timer-for-timing-code-fragments.html
311|article/Serialization-in-C.html
311|kb/serialization-in-c.html
312|article/Pickling-serialization-in-Python.html
312|kb/pickling-serialization-in-python.html
323|article/On-The-22-Laws-Of-Marketing.html
323|blog/2002/06/17/on-the-22-laws-of-marketing.html
333|article/Redefining-Professionalism-for-Software-Engineer.html
333|blog/2002/07/01/redefining-professionalism-for-software-engineer.html
335|article/Laws-of-marketing-2-category.html
335|blog/2002/07/05/laws-of-marketing-2-category.html
337|article/Laws-of-marketing-3-mind.html
337|blog/2002/07/05/laws-of-marketing-3-mind.html
338|article/Laws-of-marketing-4-perception.html
338|blog/2002/07/06/laws-of-marketing-4-perception.html
341|article/Laws-of-marketing-7-ladder.html
341|blog/2002/07/07/laws-of-marketing-7-ladder.html
342|article/Laws-of-marketing-8-duality.html
342|blog/2002/07/08/laws-of-marketing-8-duality.html
349|article/Laws-of-marketing-14-attributes.html
349|blog/2002/07/12/laws-of-marketing-14-attributes.html
350|article/Fine-interview-with-Marcelo-Tosatti.html
350|blog/2002/07/12/fine-interview-with-marcelo-tosatti.html
351|article/Laws-of-marketing-15-candor.html
351|blog/2002/07/12/laws-of-marketing-15-candor.html
360|article/Bugs-and-eyeballs.html
360|blog/2002/07/19/bugs-and-eyeballs.html
362|article/Open-Source-is-Philanthropy.html
362|blog/2002/07/23/open-source-is-philanthropy.html
367|article/Wozniaks-speech.html
367|blog/2002/08/01/wozniaks-speech.html
379|article/How-to-be-a-leader-in-your-field.html
379|blog/2002/08/11/how-to-be-a-leader-in-your-field.html
381|article/The-value-of-programming.html
381|blog/2002/08/12/the-value-of-programming.html
393|article/Information-business-as-a-relationship.html
393|blog/2002/08/27/information-business-as-a-relationship.html
402|article/WinAmp-3.html
402|blog/2002/09/02/winamp-3.html
404|article/Blog-your-resume.html
404|blog/2002/09/03/blog-your-resume.html
406|article/Interview-with-MicroStrategy-CEO.html
406|blog/2002/09/04/interview-with-microstrategy-ceo.html
416|article/You-wont-make-money-blogging.html
416|blog/2002/09/09/you-wont-make-money-blogging.html
430|article/Those-are-the-good-times.html
430|blog/2002/09/17/those-are-the-good-times.html
450|article/Show-me-the-code.html
450|blog/2002/09/28/show-me-the-code.html
460|article/Platform-Leadership.html
460|blog/2002/10/05/platform-leadership.html
462|blog/2002/10/10/slate-knows.html
469|article/Joel-man-of-his-word.html
469|blog/2002/10/20/joel-man-of-his-word.html
473|article/Open-source-lesson-from-a-stripper.html
473|blog/2002/10/27/open-source-lesson-from-a-stripper.html
476|article/How-to-sell-software.html
476|blog/2002/11/05/how-to-sell-software.html
492|article/The-ghost-of-ArsDigita.html
492|blog/2002/12/19/the-ghost-of-arsdigita.html
494|article/Your-life.html
494|blog/2003/01/05/your-life.html
495|article/Catch-me-if-you-can.html
495|blog/2003/01/05/catch-me-if-you-can.html
511|article/Successful-telecommuting.html
511|blog/2003/01/17/successful-telecommuting.html
531|article/SICP-lectures-available-on-line.html
531|blog/2003/01/31/sicp-lectures-available-on-line.html
547|article/Creative-commons-presentation.html
547|blog/2003/02/17/creative-commons-presentation.html
548|article/Inspiring-marketing-article.html
548|blog/2003/02/17/inspiring-marketing-article.html
569|article/An-old-ad-for-a-job-at-Microsoft.html
569|blog/2003/03/14/an-old-ad-for-a-job-at-microsoft.html
571|article/Outsourcing.html
571|blog/2003/03/22/outsourcing.html
578|article/Asking-the-right-question-about-language-design.html
578|blog/2003/04/01/asking-the-right-question-about-language-design.html
580|article/Abut-Face-second-edition.html
580|blog/2003/04/03/abut-face-second-edition.html
596|article/Are-Microsoft-products-any-good.html
596|blog/2003/04/22/are-microsoft-products-any-good.html
631|article/Carmack-on-creativity.html
631|blog/2003/05/10/carmack-on-creativity.html
649|article/Is-software-industry-a-place-to-be-Greenspun-per.html
649|blog/2003/05/31/is-software-industry-a-place-to-be-greenspun-per.html
657|article/On-difference-between-amateur-and-professional-s.html
657|blog/2003/06/11/on-difference-between-amateur-and-professional-s.html
665|article/Good-software-bad-buying-experience.html
665|blog/2003/06/26/good-software-bad-buying-experience.html
667|article/cmdexe-replacement-for-Windows.html
667|blog/2003/07/01/cmd-exe-replacement-for-windows.html
673|article/As-we-may-think.html
673|blog/2003/07/14/as-we-may-think.html
674|article/Usability-Heuristics-for-Rich-Internet-Applicati.html
674|blog/2003/07/16/usability-heuristics-for-rich-internet-applicati.html
683|article/Better-selling-through-a-web-site.html
683|blog/2003/08/15/better-selling-through-a-web-site.html
688|article/Popular-fallacies.html
688|blog/2003/08/20/popular-fallacies.html
695|article/Shirky-on-Wikis.html
695|blog/2003/08/27/shirky-on-wikis.html
729|blog/2003/10/20/marketing-and-sharware-articles.html
731|article/How-to-make-money-developing-Mac-apps.html
731|blog/2003/11/12/how-to-make-money-developing-mac-apps.html
732|article/Watch-TV-on-the-internet.html
732|blog/2003/11/12/watch-tv-on-the-internet.html
733|article/Skype-as-an-example-of-changing-nature-of-social.html
733|blog/2003/11/14/skype-as-an-example-of-changing-nature-of-social.html
741|article/Royalties-in-game-buisness.html
741|blog/2003/12/02/royalties-in-game-buisness.html
743|article/The-story-of-Photoshop.html
743|blog/2003/12/05/the-story-of-photoshop.html
747|article/Making-money-with-shareware-software.html
747|blog/2003/12/07/making-money-on-shareware.html
752|article/What-people-want.html
752|blog/2003/12/23/what-people-want.html
782|article/Startup-A-Silicon-Valley-Adventure-book-review.html
782|blog/2004/05/29/startup-a-silicon-valley-adventure-book-review.html
795|article/Patterns-in-interaction-design-web-and-gui-desig.html
795|blog/2004/06/02/patterns-in-interaction-design-web-and-gui-desig.html
802|article/scdiff-show-diffs-of-local-changes-in-CVS-or-Sub.html
802|blog/2004/06/04/scdiff-show-diffs-of-local-changes-in-cvs-or-sub.html
805|article/NET-Framework-bootstrapper.html
805|blog/2004/06/05/net-framework-bootstrapper.html
821|article/A-tip-from-Getting-things-done.html
821|blog/2004/06/10/a-tip-from-getting-things-done.html
826|article/Productivity-tips.html
826|blog/2004/06/12/more-productivity-tips.html
832|article/wTail-release.html
832|blog/2004/06/14/wtail-release.html
873|article/University-of-Washington-on-line-videos.html
873|blog/2004/10/22/university-of-washington-on-line-videos.html
878|article/Google-ultimate-hypocrite.html
878|blog/2004/12/25/google-ultimate-hypocrite.html
880|article/Font-Vera-Sans-Mono-recommended-for-programmers.html
880|blog/2004/12/25/font-vera-sans-mono-recommended-for-programmers.html
887|article/Counterpost-to-a-counterpost.html
887|blog/2004/12/31/counterpost-to-a-counterpost.html
891|article/Google-what-kind-of-a-giant-they-are.html
891|blog/2005/01/02/google-what-kind-of-a-giant-they-are.html
892|article/Google-saga-episode-205.html
892|blog/2005/01/02/google-saga-episode-205.html
895|article/Subversion-with-SSH-on-Windows-tip.html
895|blog/2005/02/09/subversion-with-ssh-on-windows-tip.html
898|article/How-to-delete-a-file-you-get-from-urlliburlretri.html
898|blog/2005/05/05/how-to-delete-a-file-you-get-from-urllib-urlretr.html
899|article/Backpack-observations.html
899|blog/2005/05/06/backpack-observations.html
900|article/musikCube-nice-mp3-player.html
900|blog/2005/05/10/musikcube-nice-mp3-player.html
912|article/Open-source-and-windows.html
912|blog/2005/10/13/open-source-and-windows.html
913|article/Interesting-Dave-Winer-interview.html
913|blog/2005/10/17/interesting-dave-winer-interview.html
919|article/Unsolved-source-control-problems.html
919|blog/2005/10/25/unsolved-source-control-problems.html
921|article/Petzold-on-Visual-Studio-and-mind-corruption.html
921|blog/2005/10/26/petzold-on-visual-studio-and-mind-corruption.html
928|article/Debugging-adventure.html
928|blog/2006/01/13/debugging-adventure.html
943|article/Sumatra-PDF-is-born.html
943|blog/2006/06/03/sumatra-pdf-is-born.html
944|article/Short-tutorial-on-svn-propset-for-svnexternals-p.html
944|blog/2006/06/07/short-tutorial-on-svn-propset-for-svn-externals.html
946|article/Sumatra-PDF-02-released.html
946|blog/2006/08/07/sumatra-pdf-0-2-released.html
949|article/Where-do-bugs-come-from-and-how-to-avoid-them.html
949|blog/2006/08/12/where-do-bugs-come-from-and-how-to-avoid-them.html
950|article/Performance-optimization-story.html
950|blog/2006/08/14/performance-optimization-story.html
952|article/Order-of-include-headers-in-CC.html
952|blog/2006/08/15/order-of-include-headers-in-cc.html
953|article/Paradox-of-bad-comments.html
953|blog/2006/08/16/paradox-of-bad-comments.html
954|article/A-simple-catchpa-scheme.html
954|blog/2006/08/17/a-simple-catchpa-scheme.html
955|article/What-I-love-about-Google-open-source-project-hos.html
955|blog/2006/08/20/what-i-love-about-google-open-source-project-hos.html
962|article/Sumatra-PDF-03-released.html
962|blog/2006/11/26/sumatra-pdf-0-3-released.html
967|article/SumatraPDF-05-released.html
967|blog/2007/03/05/sumatrapdf-0-5-released.html
968|article/2-great-books-and-one-not-so-great.html
968|blog/2007/04/12/2-great-books-and-one-not-so-great.html
969|article/Few-things-Ive-learned-when-writing-Sumatra-PDF.html
969|blog/2007/04/14/few-things-ive-learned-when-writing-sumatra-pdf.html
970|article/A-debugging-story.html
970|blog/2007/04/29/a-debugging-story.html
971|article/SumatraPDF-06-released.html
971|blog/2007/04/29/sumatrapdf-0-6-released.html
976|article/Sumatra-PDF-07-released.html
976|blog/2007/07/30/sumatra-pdf-0-7-released.html
978|article/Sumatra-08-released.html
978|blog/2008/01/04/sumatra-0-8-released.html
980|article/Logging-in-WinDBG.html
980|blog/2008/01/07/logging-in-windbg.html
994|article/Remapping-Page-Up-and-Page-Down-on-Mac-to-move-a.html
994|blog/2008/04/17/remapping-page-up-and-page-down-on-mac-to-move-a.html
995|article/_NT_SYMBOL_PATH-considered-harmful.html
995|blog/2008/04/18/nt-symbol-path-considered-harmful.html
996|article/Extreme-size-optimization-in-C-and-C.html
996|blog/2008/05/20/extreme-size-optimization-in-c-and-c.html
998|article/Google-App-Engine-tip.html
998|blog/2008/07/05/google-app-engine-tip.html
999|article/Announcing-fofou-forum-software-for-Google-App-E.html
999|blog/2008/07/06/announcing-fofou-forum-software-for-google-app-e.html
1010|article/making-unix-user-a-sudoer.html
1012|article/Python-static-code-checkers.html
1020|article/screen-basics.html
1025|article/Those-who-adapt-survive.html
1034|article/gcc.html
1043|article/valgrind-basics-1.html
1055|article/Faster-metabolism.html
1076|article/A-way-to-simulate-various-network-conditions.html
1085|article/International-bank-recommendations.html
1096|article/Windbg-reference.html
1122|article/enabling-coredumps.html
1169|article/DHL-in-San-Francisco.html
1189|article/Results-of-tweaking-compiler-flags-before-09-rel.html
1203|article/Objective-C-patterns.html
1231|article/Reverse-DNS-lookup.html
1243|article/variadic-macros-in-msvc.html
1253|article/Variadic-Macros-C.html
1286|article/Sane-include-hierarchy-for-C-and-C.html
1286|kb/sane-include-hierarchy-for-c-and-c.html
1289|article/Gdb-basics.html
1289|kb/gdb-basics.html
1291|article/tar-basics.html
1291|kb/tar-basics.html
1292|article/What-makes-a-CD-bootable.html
1292|kb/what-makes-a-cd-bootable.html
1293|article/C-portability-notes.html
1293|kb/c-portability-notes.html
1294|article/Embedding-binary-resources-on-Windows.html
1294|kb/embedding-binary-resources-on-windows.html
1307|article/Getting-user-specific-application-data-directory.html
1307|kb/getting-user-specific-application-data-directory-for-.net-winforms-apps.html
1308|article/Local-DNS-modifications-on-Windows-etchosts-equi.html
1308|kb/local-dns-modifications-on-windows-etchosts-equivalent.html
1309|article/Accurate-timers-on-Windows.html
1309|kb/accurate-timers-on-windows.html
1329|article/Laws-of-marketing-1-leadership.html
1329|blog/2002/07/02/laws-of-marketing-1-leadership.html
1331|article/Laws-of-marketing-5-focus.html
1331|blog/2002/07/06/laws-of-marketing-5-focus.html
1332|article/Laws-of-marketing-6-exclusivity.html
1332|blog/2002/07/07/laws-of-marketing-6-exclusivity.html
1335|article/Laws-of-marketing-9-opposite.html
1335|blog/2002/07/10/laws-of-marketing-9-opposite.html
1336|article/Laws-of-marketing-10-division.html
1336|blog/2002/07/10/laws-of-marketing-10-division.html
1337|article/Laws-of-marketing-11-perspective.html
1337|blog/2002/07/11/laws-of-marketing-11-perspective.html
1338|article/Laws-of-marketing-12-line-extension.html
1338|blog/2002/07/11/laws-of-marketing-12-line-extension.html
1339|article/Laws-of-marketing-13-sacrifice.html
1339|blog/2002/07/11/laws-of-marketing-13-sacrifice.html
1342|article/Laws-of-marketing-16-singularity.html
1342|blog/2002/07/13/laws-of-marketing-16-singularity.html
1343|article/Laws-of-marketing-17-unpredictability.html
1343|blog/2002/07/14/laws-of-marketing-17-unpredictability.html
1344|article/Laws-of-marketing-18-success.html
1344|blog/2002/07/14/laws-of-marketing-18-success.html
1345|article/Laws-of-marketing-19-failure.html
1345|blog/2002/07/15/laws-of-marketing-19-failure.html
1346|article/Laws-of-marketing-20-hype.html
1346|blog/2002/07/16/laws-of-marketing-20-hype.html
1347|article/Laws-of-marketing-21-acceleration.html
1347|blog/2002/07/16/laws-of-marketing-21-acceleration.html
1348|article/Laws-of-marketing-22-resources.html
1348|blog/2002/07/17/laws-of-marketing-22-resources.html
1349|article/You-and-your-research.html
1349|blog/2002/07/17/you-and-your-research.html
1355|article/Principle-of-good-design-discoverability.html
1355|blog/2002/07/26/principle-of-good-design-discoverability.html
1361|article/C-Interfaces-and-Implementations.html
1361|blog/2002/08/03/c-interfaces-and-implementations.html
1364|article/Stuff-costs-more-than-you-think.html
1364|blog/2002/08/05/stuff-costs-more-than-you-think.html
1381|article/On-writing-well.html
1381|blog/2002/08/21/on-writing-well.html
1389|article/The-future-is-here-its-just-not-evenly-distribut.html
1389|blog/2002/08/28/the-future-is-here-its-just-not-evenly-distribut.html
1395|article/Quote-from-Net-Words.html
1395|blog/2002/09/03/quote-from-net-words.html
1399|article/The-stupidest-thing-a-software-company-can-do.html
1399|blog/2002/09/04/the-stupidest-thing-a-software-company-can-do.html
1408|article/A-lesson-in-marketing-needed.html
1408|blog/2002/09/11/a-lesson-in-marketing-needed.html
1414|article/Great-business-without-innovation.html
1414|blog/2002/09/16/great-business-without-innovation.html
1418|article/Youll-have-a-job.html
1418|blog/2002/09/17/youll-have-a-job.html
1445|article/High-level-not-so-good.html
1445|blog/2002/10/06/high-level-not-so-good.html
1450|article/Profitable-open-source-business.html
1450|blog/2002/10/13/profitable-open-source-business.html
1463|article/How-to-refuse-features.html
1463|blog/2002/11/06/how-to-refuse-features.html
1465|article/LL2-webcast.html
1465|blog/2002/11/10/ll2-webcast.html
1467|article/Good-programming-practices.html
1467|blog/2002/11/17/good-programming-practices.html
1476|article/Blown-to-bits.html
1476|blog/2002/12/14/blown-to-bits.html
1478|article/Recruitment-is-like-dating.html
1478|blog/2002/12/16/recruitment-is-like-dating.html
1481|article/Selling-Microsoft.html
1481|blog/2002/12/19/selling-microsoft.html
1508|article/Source-Insight-35.html
1508|blog/2003/01/20/source-insight-3-5.html
1530|article/Old-ArsDigita-content.html
1530|blog/2003/01/31/old-arsdigita-content.html
1599|article/Do-you-read-the-old-papers.html
1599|blog/2003/04/26/do-you-read-the-old-papers.html
1640|article/Given-enough-eyeballs-make-all-bugs-shallow.html
1640|blog/2003/06/05/given-enough-eyeballs-make-all-bugs-shallow.html
1644|article/Writing-to-sell.html
1644|blog/2003/06/13/writing-to-sell.html
1647|article/Another-ArsDigita-story.html
1647|blog/2003/06/20/another-arsdigita-story.html
1649|article/My-future-is-so-bright-that-Ill-need-to-wear-sun.html
1649|blog/2003/06/23/my-future-is-so-bright-that-ill-need-to-wear-sun.html
1650|article/Why-consistency-is-important-in-software-design.html
1650|blog/2003/06/25/why-consistency-is-important-in-software-design.html
1653|article/Software-can-always-be-better.html
1653|blog/2003/06/28/software-can-always-be-better.html
1654|article/Programmers-dont-steal-enough.html
1654|blog/2003/06/30/programmers-dont-steal-enough.html
1656|article/OReilly-on-software.html
1656|blog/2003/07/03/oreilly-on-software.html
1659|article/How-much-can-you-make-writing-computer-books.html
1659|blog/2003/07/09/how-much-can-you-make-writing-computer-books.html
1663|article/Memex-sue-me-please-device.html
1663|blog/2003/07/15/memex-sue-me-please-device.html
1670|article/Century-dictionary-on-line.html
1670|blog/2003/07/23/century-dictionary-on-line.html
1672|article/Lucene-for-searching-source-code.html
1672|blog/2003/07/24/lucene-for-searching-source-code.html
1694|article/Not-as-happy-as-you-thought-you-will-be.html
1694|blog/2003/09/08/not-as-happy-as-you-thought-you-will-be.html
1697|article/Critical-reading-skills.html
1697|blog/2003/09/10/critical-reading-skills.html
1719|article/A-shameless-rip-off-or-what-did-you-expect.html
1719|blog/2003/10/14/a-shameless-rip-off-or-what-did-you-expect.html
1732|article/C-programming-tips-from-Rob-Pike.html
1732|blog/2003/11/19/c-programming-tips-from-rob-pike.html
1744|article/Myths-Open-Source-Developers-Tell-Ourselves.html
1744|blog/2003/12/18/myths-open-source-developers-tell-ourselves.html
1787|article/Web-writing-that-works.html
1787|blog/2004/06/02/web-writing-that-works.html
1790|article/Blogs-should-always-provide-previous-posts-butto.html
1790|blog/2004/06/02/blogs-should-always-provide-previous-posts-butto.html
1841|article/Microsoft-leading-the-way-with-open-bug-database.html
1841|blog/2004/06/30/microsoft-leading-the-way-with-open-bug-database.html
1849|article/Review-of-Hot-text-web-writing-that-works.html
1849|blog/2004/07/15/review-of-hot-text-web-writing-that-works.html
1853|article/Dont-use-0-instead-of-NULL.html
1853|blog/2004/07/22/dont-use-0-instead-of-null.html
1859|article/A-collaborative-text-editor-for-Windows.html
1859|blog/2004/08/30/a-collaborative-text-editor-for-windows.html
1860|article/DocSynch-multi-editor-plugin-for-collaborative-t.html
1860|blog/2004/08/31/docsynch-multi-editor-plugin-for-collaborative-t.html
1863|article/scdiff-03-released.html
1863|blog/2004/10/03/scdiff-0-3-released.html
1866|article/Alan-Cox-on-writing-better-software.html
1866|blog/2004/10/09/alan-cox-on-writing-better-software.html
1877|article/Recovering-data-from-formatted-drives.html
1877|blog/2004/12/13/recovering-data-from-formatted-drives.html
1882|article/GPL-3-anti-patent-virus.html
1882|blog/2004/12/27/gpl-3-anti-patent-virus.html
1884|article/Google-we-take-it-all-give-nothing-back.html
1884|blog/2004/12/30/google-we-take-it-all-give-nothing-back.html
1887|article/2005-prediction-the-rise-of-anonymous-p2p.html
1887|blog/2004/12/31/2005-prediction-the-rise-of-anonymous-p2p.html
1889|article/Bad-Google-the-fallout.html
1889|blog/2004/12/31/bad-google-the-fallout.html
1890|article/Google-comments-on-comments.html
1890|blog/2004/12/31/google-comments-on-comments.html
1901|article/Deep-indentation-vs-flat.html
1901|blog/2005/07/10/deep-indentation-vs-flat.html
1902|article/VirtualEarth-vs-Google-Maps-not-hitting-the-high.html
1902|blog/2005/07/25/virtualearth-vs-google-maps-not-hitting-the-high.html
1903|article/LonghornVista-fonts.html
1903|blog/2005/07/29/longhornvista-fonts.html
1919|article/Rich-client-is-here.html
1919|blog/2005/10/25/rich-client-is-here.html
1923|article/Code-name-Monad-and-the-value-of-different-persp.html
1923|blog/2005/10/26/code-name-monad-and-the-value-of-different-persp.html
1925|article/A-book-to-read-talks-to-listen-to.html
1925|blog/2005/10/27/a-book-to-read-talks-to-listen-to.html
1931|article/UI-design-tip-icons-are-not-enough.html
1931|blog/2005/11/02/ui-design-tip-icons-are-not-enough.html
1935|article/Another-lesson-in-entrepreneurship.html
1935|blog/2005/12/28/another-lesson-in-entrepreneurship.html
1937|article/Pawn-yet-another-embedable-language.html
1937|blog/2006/01/14/yet-another-embedable-language.html
1943|article/Digg-and-the-craft-of-catchy-headlines.html
1943|blog/2006/03/11/digg-and-the-craft-of-catchy-headlines.html
1944|article/Document-your-software.html
1944|blog/2006/03/12/document-your-software.html
1947|article/Designing-web-forums-software.html
1947|blog/2006/03/18/designing-web-forums-software.html
1954|article/Python-id3-library.html
1954|blog/2006/04/11/python-id3-library.html
1957|article/php_mysqldll-not-loading-in-PHP-514-and-Apache-2.html
1957|blog/2006/08/07/php-mysql-dll-not-loading-in-php-5-1-4-and-apach.html
1958|article/The-missing-msvcr80dll-story.html
1958|blog/2006/08/07/the-missing-msvcr80-dll-story.html
1964|article/Deeply-nested-if-statements.html
1964|blog/2006/08/22/deeply-nested-if-statements.html
1966|article/On-how-I-improved-Sumatra-performance-by-60.html
1966|blog/2006/09/03/on-how-i-improved-sumatra-performance-by-~60.html
1969|article/Navigating-source-code-in-large-programs.html
1969|blog/2006/09/21/navigating-source-code-in-large-programs.html
1970|article/Talk-on-designing-good-APIs.html
1970|blog/2006/11/22/talk-on-designing-good-apis.html
1971|article/Programmers-are-silver-bullets-or-after-all-this.html
1971|blog/2006/12/08/programmers-are-silver-bullets-or-after-all-this.html
1973|article/memset-considered-harmful.html
1973|blog/2007/02/16/memset-considered-harmful.html
1974|article/SumatraPDF-04-released.html
1974|blog/2007/02/20/sumatrapdf-0-4-released.html
1982|article/Merge-tools-showdown.html
1982|blog/2007/07/30/merge-tools-showdown.html
1986|article/Rebol-vs-Shoes.html
1986|blog/2008/01/09/rebol-vs-shoes.html
1987|article/Too-much-oo.html
1987|blog/2008/01/11/too-much-oo.html
1989|article/gflags-a-debugging-story.html
1989|blog/2008/04/07/gflags-a-debugging-story.html
1990|article/Google-App-Engine-the-first-Internet-operating-s.html
1990|blog/2008/04/08/google-app-engine-the-first-internet-operating-s.html
1999|article/SumatraPDF-081.html
1999|blog/2008/05/29/sumatrapdf-0-8-1.html
2005|article/SumatraPDF-09-released.html
2005|blog/2008/08/11/sumatrapdf-0-9-released.html
2006|article/SumatraPDF-091-released.html
2006|blog/2008/08/24/sumatrapdf-0-9-1-released.html
2007|article/SumatraPDF-093-released.html
2007|blog/2008/10/02/sumatrapdf-0-9-3-released.html
2015|article/Exporting-data-from-EverNote.html
2017|article/BitTorrent-based-large-file-distribution-for-HTT.html
2075|article/Experience-with-using-Rietveld-for-code-reviews.html
2101|article/Cocoa-source-code-and-tutorials.html
3002|article/realloc-on-Windows-vs-Linux-1.html
3002|blog/2008/07/27/realloc-on-windows-vs-linux.html
3018|article/App-Engine-as-generic-web-host.html
3081|article/NSCopying-NSMutableCopying-or-NSCoding.html
3082|article/Fonts-on-windows.html
3089|article/Profiling-tools-for-CC-on-windows-mac-and-linux.html
7045|article/Essential-software.html
7051|article/ssh-tips.html
8001|article/Where-do-bugs-come-from.html
8003|article/Resources-related-to-implementing-programming-la.html
8047|article/Summary-of-David-Ditzel-talk-on-binary-translati.html
8074|article/How-content-based-addressing-can-help-web-perfor.html
8077|article/Interesting-win32-source-code.html
12012|article/Compacting-s3-aws-logs.html
13010|article/Parsing-s3-log-files-in-python.html
14005|article/setting-up-s3-logging.html
14007|article/Forcing-basic-http-authentication-for-HttpWebReq.html
15004|article/scdiff-update-Windows-gitsubversioncvs-gui-diff-.html
18003|article/15minutes-a-simple-productivity-tool.html
19006|article/Setting-unicode-rtf-text-in-rich-edit-control.html
20003|article/Accessing-Mac-file-shares-from-Windows-7.html
21002|article/Automatic-Java-to-C-conversion-experience-using-.html
25011|article/Network-drives-net-security-and-virtualbox.html
34002|article/15minutes-for-mac-now-available.html
35002|article/Sumatra-094-release.html
37001|article/We-need-Visual-Ack.html
41001|article/Unicode-problem-with-firstof-in-appengineDjango.html
44002|article/15minutes-for-mac-updated.html
45002|article/Web-server-in-C.html
47001|article/SumatraPDF-10-released.html
48002|article/15minutes-11-for-windows.html
50001|article/Drobo-Dashboard-and-mysterious-mac-slowdowns.html
55002|article/You-have-to-implement-to-understand.html
55004|article/Best-captcha-is-exotic-captcha.html
57003|article/VisualAck-032-released.html
59001|article/VisualAck-033-released.html
81001|article/e-books-economics.html
93002|article/uISV-stories.html
94001|article/Productivity-ideas.html
98001|article/Things-Ive-learned-this-week.html
128001|article/SumatraPDF-11-release.html
134001|article/Summary-of-talk-on-continuous-deployment.html
148001|article/How-to-accept-online-payments.html
204001|article/Go-vs-Python-for-a-simple-web-server.html
212001|article/Software-licensing-scheme.html
229002|article/Converting-PartCover-results-to-html.html
238002|article/Hiding-duplicate-content-from-your-site-via-robo.html
254001|article/Searching-for-available-DBA-name-in-San-Francisc.html
256002|article/Comparing-program-versions-in-C-and-Python.html
266001|article/Tools-that-find-bugs-in-c-and-c-code-via-static-.html
286001|article/Introduction-to-PartCover-a-short-manual.html
314001|article/SEO-is-harder-than-you-think.html
319001|article/Why-you-shouldnt-write-Mac-programs-in-QT.html
322001|article/Beware-spurious-charges-when-buying-from-Paralle.html
330001|article/Marketing-lessons-from-WebP-launch.html
331001|article/Startup-management-lessons-from-The-Social-Netwo.html
334001|article/Value-your-time.html
336001|article/Simple-duplicate-post-detection-for-your-blog-fo.html
338001|article/8-habits-for-becoming-a-better-programmer.html
340001|article/Using-averages-a-common-performance-measurement-.html
342001|article/SumatraPDF-12-released.html
346001|article/Which-technology-for-writing-desktop-software.html
`

var redirects = [][]string{
	{"/index.html", "/"},
	{"/blog", "/"},
	{"/blog/", "/"},
	{"/kb/serialization-in-c#.html", "/article/Serialization-in-C.html"},
	{"/extremeoptimizations", "/extremeoptimizations/index.html"},
	{"/extremeoptimizations/", "/extremeoptimizations/index.html"},
	{"/feed/rss2/atom.xml", "/atom.xml"},
	{"/feed/rss2/", "/atom.xml"},
	{"/feed/rss2", "/atom.xml"},
	{"/feed/", "/atom.xml"},
	{"/feed", "/atom.xml"},
	{"/feedburner.xml", "/atom.xml"},
	{"/articles/cocoa-objectivec-reference.html", "/articles/cocoa-reference.html"},
	{"/forum_sumatra", "https://forum.sumatrapdfreader.org/"},
	{"/google6dba371684d43cd6.html", "/static/google6dba371684d43cd6.html"},
	{"/software/15minutes/index.html", "/software/15minutes.html"},
	{"/software/15minutes/", "/software/15minutes.html"},
	{"/software/fofou", "/software/fofou/index.html"},
	{"/software/dbhero", "/software/dbhero/index.html"},
	{"/software/patheditor", "/software/patheditor/for-windows.html"},
	{"/software/patheditor/", "/software/patheditor/for-windows.html"},
	{"/software/scdiff/", "/software/scdiff.html"},
	{"/software/scdiff/index.html", "/software/scdiff.html"},
	{"/software/sumatra", "https://www.sumatrapdfreader.org/free-pdf-reader.html"},
	{"/software/sumatrapdf", "https://www.sumatrapdfreader.org/free-pdf-reader.html"},
	{"/software/sumatrapdf/", "https://www.sumatrapdfreader.org/free-pdf-reader.html"},
	{"/software/sumatrapdf/index.html", "https://www.sumatrapdfreader.org/free-pdf-reader.html"},
	{"/software/sumatrapdf/download.html", "https://www.sumatrapdfreader.org/download-free-pdf-viewer.html"},
	{"/software/sumatrapdf/prerelase.html", "https://www.sumatrapdfreader.org/prerelease.html"},
	{"/free-pdf-reader.html", "https://www.sumatrapdfreader.org/free-pdf-reader.html"},
	{"/software/volante", "/software/volante/database.html"},
	{"/software/volante/", "/software/volante/database.html"},
	{"/software/volante/index.html", "/software/volante/database.html"},
	{"/software/fotofi", "/software/fotofi/free-stock-photos.html"},
	{"/software/fotofi/", "/software/fotofi/free-stock-photos.html"},
	{"/software/fotofi/index.html", "/software/fotofi/free-stock-photos.html"},
	{"/static/software.html", "/software/index.html"},
	{"/static/krzysztof.html", "/resume.html"},
	{"/static/resume.html", "/resume.html"},
}

var articleRedirects = make(map[string]string)

func readRedirects() {
	d := []byte(articleRedirectsTxt)
	lines := bytes.Split(d, []byte{'\n'})
	for _, l := range lines {
		if 0 == len(l) {
			continue
		}
		parts := strings.Split(string(l), "|")
		panicIf(len(parts) != 2, "malformed article_redirects.txt, len(parts) = %d (!2)", len(parts))
		idStr := parts[0]
		url := strings.TrimSpace(parts[1])
		idNum, err := strconv.Atoi(idStr)
		panicIf(err != nil, "malformed line in article_redirects.txt. Line:\n%s\nError: %s\n", l, err)
		id := u.EncodeBase64(idNum)
		a := getArticleByID(id)
		if a != nil {
			articleRedirects[url] = id
			continue
		}
		//fmt.Printf("skipping redirect '%s' because article with id %d no longer present\n", string(l), id)
	}
}

var (
	netlifyRedirects []*netlifyRedirect
)

type netlifyRedirect struct {
	from string
	to   string
	// valid code is 301, 302, 200, 404
	code int
}

func netlifyAddRedirect(from, to string, code int) {
	r := netlifyRedirect{
		from: from,
		to:   to,
		code: code,
	}
	netlifyRedirects = append(netlifyRedirects, &r)
}

func netlifyAddRewrite(from, to string) {
	netlifyAddRedirect(from, to, 200)
}

func netflifyAddTempRedirect(from, to string) {
	netlifyAddRedirect(from, to, 302)
}

func netflifyAddPermRedirect(from, to string) {
	netlifyAddRedirect(from, to, 301)
}

func netlifyAddStaticRedirects() {
	for _, redirect := range redirects {
		from := redirect[0]
		to := redirect[1]
		netflifyAddTempRedirect(from, to)
	}
}

func netlifyAddArticleRedirects() {
	for from, articleID := range articleRedirects {
		from = "/" + from
		article := getArticleByID(articleID)
		panicIf(article == nil, "didn't find article for id '%s'", articleID)
		to := article.URL()
		netflifyAddTempRedirect(from, to) // TODO: change to permanent
	}

}

// redirect /article/:id/* => /article/:id/pretty-title
const netlifyRedirectsProlog = `/article/:id/*	/article/:id.html	200
`

func netlifyWriteRedirects() {
	buf := bytes.NewBufferString(netlifyRedirectsProlog)
	for _, r := range netlifyRedirects {
		s := fmt.Sprintf("%s\t%s\t%d\n", r.from, r.to, r.code)
		buf.WriteString(s)
	}
	netlifyWriteFile("_redirects", buf.Bytes())
}

// https://caddyserver.com/tutorial/caddyfile
// redirect /article/:id/* => /article/:id/pretty-title
var caddyProlog = `localhost:8080
root netlify_static
errors stdout
log stdout

rewrite / {
	r  ^/article/(.*)/.*$
	to /article/{1}.html
}

`

func isRewrite(r *netlifyRedirect) bool {
	return (r.code == 200) || strings.HasSuffix(r.from, "*")
}

func genCaddyRedir(r *netlifyRedirect) string {
	if r.from == "/" {
		return fmt.Sprintf("rewrite / %s\n", r.to)
	}
	if isRewrite(r) {
		// hack: caddy doesn't like `++` in from
		if strings.Contains(r.from, "++") {
			return ""
		}
		if strings.HasSuffix(r.from, "*") {
			base := strings.TrimSuffix(r.from, "*")
			to := strings.Replace(r.to, ":splat", "{1}", -1)
			return fmt.Sprintf("rewrite \"%s\" {\n    regexp (.*)\n    to %s\n}\n", base, to)
		}
		return fmt.Sprintf("rewrite \"%s\" \"%s\"\n", r.from, r.to)
	}

	return fmt.Sprintf("redir \"%s\" \"%s\" %d\n", r.from, r.to, r.code)
}

func writeCaddyConfig() {
	path := filepath.Join("Caddyfile")
	f, err := os.Create(path)
	panicIfErr(err)
	defer f.Close()

	_, err = f.Write([]byte(caddyProlog))
	panicIfErr(err)
	for _, r := range netlifyRedirects {
		s := genCaddyRedir(r)
		_, err = io.WriteString(f, s)
		panicIfErr(err)
	}
}
