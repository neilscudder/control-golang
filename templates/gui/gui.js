{{define "JS"}}
<script>var clickEventType = ((document.ontouchstart!==null)?'click':'touchstart')
var PreviousInfo
// TODO fallback to APIALT if necessary
function getURLParameter(name) {
  return decodeURIComponent((new RegExp('[?|&]' + name + '=' + '([^&;]+?)(&|#|;|$)').exec(location.search)||[,""])[1].replace(/\+/g, '%20'))||null
}

getparams = getURLParameter('APIURL')
  + "get"
  + "?KPASS=" + getURLParameter('KPASS');

function autoRefresh(id) {
  //console.log("Auto-Refresh: " + id)
  sendCmd('info')
  setTimeout(function(){ autoRefresh(id) },1000)
} 
function sendCmd(id) {
  //console.log("sendCmd: " + id)
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
    //console.log("YO" + id)
    infoDiv.classList.remove('heartbeat')
    infoDiv.classList.add('opaque')
    if (CurrentInfo !== PreviousInfo && !isEmpty(CurrentInfo)) {
      infoDiv.innerHTML = CurrentInfo
      PreviousInfo = CurrentInfo
      animatedButtonListener()
      //console.log("Different")
    } else {
      //console.log("Same")
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
      buttons[i].addEventListener(clickEventType, pusher, false)
  }
}
initialise()

</script>
{{end}}
