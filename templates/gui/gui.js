{{define "JS"}}
<script>var clickEventType = ((document.ontouchstart!==null)?'click':'touchstart')
var PreviousInfo
// TODO fallback to APIALT if necessary
function getURLParameter(name) {
  return decodeURIComponent((new RegExp('[?|&]' + name + '=' + '([^&;]+?)(&|#|;|$)').exec(location.search)||[,""])[1].replace(/\+/g, '%20'))||null
}

cmdparams = getURLParameter('APIURL')
  + "cmd" 
  + "?KPASS=" + getURLParameter('KPASS');

getparams = getURLParameter('APIURL')
  + "get"
  + "?KPASS=" + getURLParameter('KPASS');

function sendCmd(id){
  var x = document.getElementById(id)
  var xhr = new XMLHttpRequest()
  params = cmdparams + "&a=" + id
  xhr.open("GET",params,true)
  xhr.send()
  xhr.onreadystatechange = function() {
    if (xhr.status == 200 && xhr.readyState == 4 && x.classList.contains("pushed")) {
      manualRefresh('info')
      x.classList.add('released')
      x.classList.remove('pushed')
    } else if (xhr.readyState == 4 && x.classList.contains("pushed")) {
      x.classList.add('denied')
      x.classList.remove('pushed')
    } else {
      // Nothing
    }
  }
}

function autoRefresh(id) {
  var x = document.getElementById('info')
  x.classList.remove('opaque')
  x.classList.add('heartbeat')

  setTimeout(function(){ autoRefresh(id) },4000)
  var xhr = new XMLHttpRequest()
  params = getparams + "&a=" + id
  xhr.open("GET",params,true)
  xhr.send()
  xhr.onreadystatechange = function() {
    if (xhr.readyState == 4 && xhr.status == 200) {
      var CurrentInfo = xhr.responseText;
      if (CurrentInfo !== PreviousInfo && !isEmpty(CurrentInfo)) {
        var div = document.getElementById(id)
        div.innerHTML = CurrentInfo
        PreviousInfo = CurrentInfo
        animatedButtonListener()
      } 
      x.classList.remove('heartbeat')
      x.classList.add('opaque')
    } 
  } 
} 
function manualRefresh(id) {
  var x = document.getElementById('info')
  x.classList.remove('opaque')
  x.classList.add('heartbeat')

  var xhr = new XMLHttpRequest()
  params = getparams + "&a=" + id
  xhr.open("GET",params,true)
  xhr.send()
  xhr.onreadystatechange = function() {
    if (xhr.readyState == 4 && xhr.status == 200) {
      var CurrentInfo = xhr.responseText;
      if (CurrentInfo !== PreviousInfo && !isEmpty(CurrentInfo)) {
        var div = document.getElementById(id)
        div.innerHTML = CurrentInfo
        PreviousInfo = CurrentInfo
        animatedButtonListener()
      } 
      x.classList.remove('heartbeat')
      x.classList.add('opaque')
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

//
// LISTENERS
//
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
