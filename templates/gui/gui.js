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

function autoRefresh(id,interval) {
  if (AutoToggle){ sendCmd(id) }
  setTimeout(function(){ autoRefresh(id,interval) },interval)
} 
//Finds y value of given object
function findPos(element) {
    var yPosition = 0;
    while(element) {
      yPosition += (element.offsetTop - element.scrollTop + element.clientTop);
      element = element.offsetParent;
    }
    return [ yPosition - 160 ]
}
function sendCmd(id) {
//  AutoToggle = false
  var xhr = new XMLHttpRequest()
  params = getparams + "&a=command&b=" + id
  xhr.addEventListener("load", transferComplete)
  xhr.open("GET",params,true)
  xhr.send()
  var infoDiv = document.getElementById('info')
  infoDiv.classList.remove('opaque')
  infoDiv.classList.add('heartbeat')
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
	scrollTo()
      }
    } else {
      var CurrentState = this.responseText
      var button = document.getElementById(id)
      var banner = document.getElementById('bannerText')
      if (CurrentState !== PreviousState && !isEmpty(CurrentState)) {
        state = JSON.parse(CurrentState)
        PreviousState = CurrentState
      	banner.innerHTML = state.Banner
    	  if (state.Random == 0) { 
    	    document.getElementById("random").style.backgroundColor = "#586e75"
          document.getElementById("random").style.opacity = "0.5"
    	  } else {
    	    document.getElementById("random").style.backgroundColor = "#839496"
          document.getElementById("random").style.opacity = "1.0"
    	  }
    	  if (state.Repeat == 0) { 
    	    document.getElementById("repeat").style.backgroundColor = "#586e75"
          document.getElementById("repeat").style.opacity = "0.5"
    	  } else {
    	    document.getElementById("repeat").style.backgroundColor = "#839496"
          document.getElementById("repeat").style.opacity = "1.0"
    	  }
    	  var playSVG = document.getElementById('playsvg')
    	  if (state.Play == 'play') {
    	    var pausePaths = '<path class=\"iconFill\" d=\"M6 19h4V5H6v14zm8-14v14h4V5h-4z\"></path><path d=\"M0 0h24v24H0z\" fill=\"none\"></path>'
    	    playsvg.innerHTML = pausePaths
    	  } else {
    	    var playPaths = '<path class=\"iconFill\" d=\"M8 5v14l11-7z\" ></path><path fill=\"none\" d=\"M0 0h24v24H0z\"></path>'
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
      var burl = '?a=browser' + "&KPASS=" + getURLParameter('KPASS') + "&APIURL=" + getURLParameter('APIURL')
;
      if (id == "browser") {window.location.replace(burl)}
      if (id == "main") {window.location.replace(burl)}
    }
  }
} 
function isEmpty(str) {
    return (!str || 0 === str.length)
}
function initialise() {
  autoRefresh('info', 1500)
  autoRefresh("state", 3000)
  animatedButtonListener()
  window.addEventListener("deviceorientation", rotHandler, true);
}
var oBuffer = 45
function rotHandler() {
	if (window.orientation != oBuffer) {
		oBuffer = window.orientation
		scrollTo()
	}
}
function scrollTo() {
	window.scroll(0,findPos(document.getElementById("scrollTo")))
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
initialise()

</script>
{{end}}
