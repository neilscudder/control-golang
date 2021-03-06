{{define "JS"}}
<script>
var ClickEventType = ((document.ontouchstart!==null)?'click':'touchstart')
var PreviousInfo
var PreviousState
var AutoToggle = true
function getURLParameter(name) {
  return decodeURIComponent((new RegExp('[?|&]' + name + '=' + '([^&;]+?)(&|#|;|$)').exec(location.search)||[,""])[1].replace(/\+/g, '%20'))||null
}

getparams = getURLParameter('APIURL')
  + "?KPASS=" + getURLParameter('KPASS');

var form = document.forms.namedItem("search");
form.addEventListener('submit', goSearch, false)

function goSearch(ev) {
  var oOutput = document.getElementById('searchResults'),
      oData = new FormData(form);

  oData.append("KPASS", getURLParameter('KPASS'));
  oData.append("a", "search");

  var xhr = new XMLHttpRequest();
  xhr.open("POST", "/", true);
  xhr.onload = function(oEvent) {
    if (xhr.status == 200) {
      oOutput.innerHTML = xhr.responseText;
      playButtonListener()
    } else {
      oOutput.innerHTML = "Error " + xhr.status + " occurred.";
    }
  };

  xhr.send(oData);
  ev.preventDefault();
}

function autoRefresh(id,interval) {
  if (AutoToggle){ sendCmd(id) }
  setTimeout(function(){ autoRefresh(id,interval) },interval)
} 
//Finds y value of given object
function findPos(obj) {
  var curtop = 0
    if (obj.offsetParent) {
      do {
        curtop += obj.offsetTop
      } while (obj = obj.offsetParent)
      return [curtop - 210]
    }
}
function sendCmd(id) {
//  AutoToggle = false
  var xhr = new XMLHttpRequest()
  params = getparams + "&a=command&b=" + id
  xhr.addEventListener("load", transferComplete)
  xhr.open("GET",params,true)
  xhr.send()
  function transferComplete() {
    AutoToggle = true
    if (id == "info") {
      var CurrentInfo = this.responseText;
      infoDiv.classList.remove('heartbeat')
      infoDiv.classList.add('opaque')
      if (CurrentInfo !== PreviousInfo && !isEmpty(CurrentInfo)) {
        infoDiv.innerHTML = CurrentInfo
        PreviousInfo = CurrentInfo
        animatedButtonListener()
	window.scroll(0,findPos(document.getElementById("scrollTo")))
      }
    } else {
      var CurrentState = this.responseText
      var button = document.getElementById(id)
      var banner = document.getElementById('BannerArea')
      if (CurrentState !== PreviousState && !isEmpty(CurrentState)) {
        state = JSON.parse(CurrentState)
        PreviousState = CurrentState
      	banner.innerHTML = state.Banner
    	  if (state.Random == 0) { 
    	    document.getElementById("random").style.backgroundColor = "#586e75"
    	  } else {
    	    document.getElementById("random").style.backgroundColor = "#839496"
    	  }
    	  if (state.Repeat == 0) { 
    	    document.getElementById("repeat").style.backgroundColor = "#586e75"
    	  } else {
    	    document.getElementById("repeat").style.backgroundColor = "#839496"
    	  }
    	  var playSVG = document.getElementById('playsvg')
    	  if (state.Play == 'play') {
    	    var pausePaths = '<path class=\"buttonPath\" d=\"M6 19h4V5H6v14zm8-14v14h4V5h-4z\"></path><path d=\"M0 0h24v24H0z\" fill=\"none\"></path>'
    	    playsvg.innerHTML = pausePaths
    	  } else {
    	    var playPaths = '<path class=\"buttonPath\" d=\"M8 5v14l11-7z\" ></path><path fill=\"none\" d=\"M0 0h24v24H0z\"></path>'
    	    playsvg.innerHTML = playPaths
    	  }
      }
      if (id != "state") { 
    		if (button.classList.contains("pushed")) {
    			button.classList.remove('pushed')
    			button.classList.add('released')
    		}
        if (id == "fw" || id == "bk") {
          setTimeout(function(){ sendCmd('info') }, 600)
          // This should have a callback to set button state to released ^^
        }
      }
      var burl = '/' + "?KPASS=" + getURLParameter('KPASS') + "&APIURL=" + getURLParameter('APIURL')
;
      if (id == "browser") {window.location.replace(burl)}
      if (id == "main") {window.location.replace(burl)}
    }
  }
} 

function playCmd(ev) {
  var x = ev.currentTarget
  var target = x.dataset.target
  var index = x.dataset.index
  var apiURL = getURLParameter('APIURL')
  apiURL = apiURL + "post"
  var oOutput = document.getElementById('searchResults'),
      oData = new FormData()
  oData.append("KPASS", getURLParameter('KPASS'))
  oData.append("a", "play")
  oData.append("b", target)
  oData.append("c", index)
  var xhr = new XMLHttpRequest()

  xhr.open("POST", apiURL, true)
  xhr.onload = function(oEvent) {
    if (xhr.status == 200) {
      var gui = '/' + "?KPASS=" + getURLParameter('KPASS') + "&APIURL=" + getURLParameter('APIURL')
      window.location.replace(gui)
    } else {
      oOutput.innerHTML = "Error " + xhr.status + " occurred."
    }
  }
  xhr.send(oData)
  ev.preventDefault()
}

function isEmpty(str) {
    return (!str || 0 === str.length)
}
function initialise() {
  autoRefresh("state", 3000)
  animatedButtonListener()
}

function pushed(id){
    document.getElementById(id).classList.add('pushed')
    document.getElementById(id).classList.remove('released')
}

function animatedButtonListener() {
  var buttons = document.getElementsByClassName("animated")
  function pusher(e){
    var id = e.currentTarget.id
    var x = document.getElementById(id)
    if (x.classList.contains("released") && id.match(/tog/g)) {
      pushed(id)
      togBrowser(id)
    } else if (x.classList.contains("released")) {
      pushed(id)
      sendCmd(id)
    }
  }
  for(i = 0; i<buttons.length; i++) {
      buttons[i].addEventListener(ClickEventType, pusher, false)
  }
}

function playButtonListener() {
  console.log("add listener")
  var buttons = document.getElementsByClassName("play")
  for(i = 0; i<buttons.length; i++) {
      buttons[i].addEventListener("touchend", playCmd, false)
  }
}

initialise()

</script>
{{end}}
