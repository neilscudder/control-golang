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
  setTimeout(function(){ autoRefresh(id) },3000)
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

      if (infoDiv.getElementsByClassName('CurrentRandom')) {
        var currnd = infoDiv.getElementsByClassName('CurrentRandom')[0].id
	if (currnd == '0') { 
      	  document.getElementById("random").style.backgroundColor = "#dc322f"
	} else {
      	  document.getElementById("random").style.backgroundColor = "#859900"
	}
      }
      if (infoDiv.getElementsByClassName('Volume')) {
        var volume = infoDiv.getElementsByClassName('Volume')[0].id
	volume = volume * 0.01
	var inverse = 1 - volume
        document.getElementById("dn").style.opacity = volume
        document.getElementById("up").style.opacity = inverse
      }
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
