function $(id) {
  if (id[0] == "#") {
    id = id.substr(1);
  }
  return document.getElementById(id);
}

var allShots = ["shot-00.png", "shot-01.png"];

function changeShot(url) {
  var el = $("main-shot");
  el.setAttribute("src", url);
  var n = allShots.length;
  for (var i = 0; i < n; i++) {
    var id = allShots[i];
    el = $(id);
    if (id == url) {
      el.classList.add("selected-img");
    } else {
      el.classList.remove("selected-img");
    }
  }
}
