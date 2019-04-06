function getEl(id) {
  if (id[0] == "#") {
    id = id.substr(1);
  }
  return document.getElementById(id);
}

var allShots = ["shot-00.png", "shot-01.png"];
var currImg = "shot-00.png";

function getImgIdx(img) {
  var n = allShots.length;
  for (var i = 0; i < n; i++) {
    if (img == allShots[i]) {
      return i;
    }
  }
  return 0;
}

function changeShot(imgUrl) {
  currImg = imgUrl;

  var el = getEl("main-shot");
  el.setAttribute("src", currImg);

  var n = allShots.length;
  var isFirstImage = false;
  var isLastImage = false;
  for (var i = 0; i < n; i++) {
    var id = allShots[i];
    el = getEl(id);
    if (id == imgUrl) {
      isFirstImage = i == 0;
      isLastImage = i == n - 1;
      el.classList.add("selected-img");
    } else {
      el.classList.remove("selected-img");
    }
  }
}

function updateDlUrl() {
  var el = getEl("#dlurl");
  el.setAttribute('href', gLatestVer.URL);
}

function advanceImage(n) {
  var idx = getImgIdx(currImg) + n;
  if (idx < 0) {
    idx = allShots.length + idx;
  } else {
    idx = idx % allShots.length;
  }
  changeShot(allShots[idx]);
}

function imgNext() {
  advanceImage(1);
}

function imgPrev() {
  advanceImage(-1);
}
