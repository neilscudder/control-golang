{{define "JS"}}
<script>
var ClickEventType = ((document.ontouchstart!==null)?'click':'touchstart')
var PreviousInfo
var AutoToggle = true
function getURLParameter(name) {
  return decodeURIComponent((new RegExp('[?|&]' + name + '=' + '([^&;]+?)(&|#|;|$)').exec(location.search)||[,""])[1].replace(/\+/g, '%20'))||null
}

getparams = getURLParameter('APIURL')
  + "get"
  + "?KPASS=" + getURLParameter('KPASS');

function autoRefresh(id) {
  if (AutoToggle){ sendCmd(id) }
  setTimeout(function(){ autoRefresh(id) },1500)
} 
function sendCmd(id) {
  AutoToggle = false
  var button = document.getElementById(id)
  var infoDiv = document.getElementById('info')
  infoDiv.classList.remove('opaque')
  infoDiv.classList.add('heartbeat')
  var xhr = new XMLHttpRequest()
  params = getparams + "&a=" + id
  xhr.addEventListener("load", transferComplete)
  xhr.open("GET",params,true)
  xhr.send()
  function transferComplete() {
    var CurrentInfo = this.responseText;
    infoDiv.classList.remove('heartbeat')
    infoDiv.classList.add('opaque')
    AutoToggle = true
    if (CurrentInfo !== PreviousInfo && !isEmpty(CurrentInfo)) {
      infoDiv.innerHTML = CurrentInfo
      PreviousInfo = CurrentInfo
      animatedButtonListener()


      if (document.getElementById("BannerText")) {
 	var BannerText = document.getElementById("BannerText")
	var banner = document.getElementById("BannerArea")
	banner.innerHTML = BannerText.dataset.content 
      }

      if (infoDiv.getElementsByClassName('CurrentRandom')) {
        var currnd = infoDiv.getElementsByClassName('CurrentRandom')[0].id
	if (currnd == '0') { 
      	  document.getElementById("random").style.backgroundColor = "#586e75"
	} else {
      	  document.getElementById("random").style.backgroundColor = "#839496"
	}
      }
      if (infoDiv.getElementsByClassName('Repeat')) {
        var cRpt = infoDiv.getElementsByClassName('Repeat')[0].id
	if (cRpt == '0') { 
      	  document.getElementById("repeat").style.backgroundColor = "#586e75"
	} else {
      	  document.getElementById("repeat").style.backgroundColor = "#839496"
	}
      }
      if (infoDiv.getElementsByClassName('PlayState')) {
        var curply = infoDiv.getElementsByClassName('PlayState')[0].id
        var playSVG = document.getElementById('playsvg')
        if (curply == 'play') {
          var pausePaths = '<path fill=\"#002B36\ "d=\"M6 19h4V5H6v14zm8-14v14h4V5h-4z\"></path><path d=\"M0 0h24v24H0z\" fill=\"none\"></path>'
          playsvg.innerHTML = pausePaths
        } else {
          var playPaths = '<path fill=\"#002B36\" d=\"M8 5v14l11-7z\" ></path><path fill=\"none\" d=\"M0 0h24v24H0z\"></path>'
          playsvg.innerHTML = playPaths
        }

      }
/*      if (infoDiv.getElementsByClassName('Volume')) {
        var volume = infoDiv.getElementsByClassName('Volume')[0].id
	volume = volume * 0.01
	var inverse = 1 - volume
	volume = volume + 0.3
	inverse = inverse + 0.3
        document.getElementById("dnsvg").style.opacity = volume
        document.getElementById("upsvg").style.opacity = inverse
      }*/
    }
    if (button.classList.contains("pushed")) {
      button.classList.remove('pushed')
      button.classList.add('released')
    }
  }
} 
function isEmpty(str) {
    return (!str || 0 === str.length)
}
function initialise() {
  var id = document.getElementsByTagName('section')[0].id
  autoRefresh(id)
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
initialise()

</script>
{{end}}
