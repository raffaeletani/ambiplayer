var counter = 0
var template = document.getElementById('playingrow')
var libtemplate = document.getElementById('libraryrow')

function setupdloadfilename(elem){
    document.getElementById('uplaodfilename').innerHTML = elem.files[0].name
}

function rangeonchange(elem){
    elem.setAttribute('data-locked','false')
}
function rangeoninput(elem){
    elem.setAttribute('data-locked','true')
}
function clearLibrary(){
    document.getElementById('library').innerHTML = ''
}
function createLibrary(json){
    if ( json == 'null'){
        return
    }
    dataArr = JSON.parse(json)
    if (dataArr.length < 1 ){
        return
    }
    if (dataArr.Lib == null){ return }
    dataArr.Lib.forEach(data => {
        container = document.getElementById('library')
        elem = document.getElementById('library-'+data.Code)
        if (elem == null) {
            elem = libtemplate.content.cloneNode(true);
            elem.querySelector('.libraryrowwrapper').setAttribute('id', 'library-'+data.Code)

            fadeintime = elem.querySelector('.fadeintimelib')
            fadeintime.value=data.DefaultFadeTime
            fadeintime.setAttribute('id', 'fadeintimelib-' + data.Code)
            fadeouttime = elem.querySelector('.fadeouttimelib')
            fadeouttime.value=data.DefaultFadeTime
            fadeouttime.setAttribute('id', 'fadeouttimelib-' + data.Code)

            startbutton = elem.querySelector('.startbutton')
            startbutton.setAttribute('hx-get', 'api/v1/start?code=' + data.Code)

            fadeinstartbutton = elem.querySelector('.fadeinstartbutton')
            fadeinstartbutton.setAttribute('hx-get', 'api/v1/start?code=' + data.Code)
            fadeinstartbutton.setAttribute('hx-include', '#fadeintimelib-' + data.Code)

            playbutton = elem.querySelector('.playbutton')
            playbutton.setAttribute('hx-get', 'api/v1/play?code=' + data.Code)

            pausebutton = elem.querySelector('.pausebutton')
            pausebutton.setAttribute('hx-get', 'api/v1/pause?code=' + data.Code)

            stopbutton = elem.querySelector('.stopbutton')
            stopbutton.setAttribute('hx-get', 'api/v1/stop?code=' + data.Code)

            fadeoutstopbutton = elem.querySelector('.fadeoutstopbutton')
            fadeoutstopbutton.setAttribute('hx-get', 'api/v1/fadeoutstop?code=' + data.Code)
            fadeoutstopbutton.setAttribute('hx-include', '#fadeouttimelib-' + data.Code)

            deletebutton = elem.querySelector('.deletebutton')
            deletebutton.setAttribute('hx-get', 'api/v1/delete?code=' + data.Code)

            
            
            

            elem.querySelector('.copycode').value = data.Code
            container.append(elem)
            elem = document.getElementById('library-'+data.Code)
            htmx.process(elem)
        }
        elem = document.getElementById('library-'+data.Code)
        elem.querySelector('.filenamedisplay').innerHTML = data.Name

        
    });
}
    
function createPlayback(json){
    if ( json == 'null'){
        return
    }
    dataArr = JSON.parse(json)
    if (dataArr.length < 1 ){
        return
    }
    if (dataArr.Play == null){ 
        elems = document.querySelectorAll("[id^='playback-']")
        elems.forEach(v => {
        v.remove()
        })
        return 
    }
    elems = document.querySelectorAll("[id^='playback-']")
    existingIds =  Array.from(elems, node => node.id.replace('playback-',''))

    dataArr.Play.forEach(data => {
        container = document.getElementById('playing')
        elem = document.getElementById('playback-'+data.Code)
        //remove current from existing
        index = existingIds.indexOf(data.Code);
        if (index > -1) { // only splice array when item is found
            existingIds.splice(index, 1); // 2nd parameter means remove one item only
        }
        if (elem == null) {
            elem = template.content.cloneNode(true);
            elem.querySelector('.rowwrapper').setAttribute('id', 'playback-'+data.Code)
            playbutton = elem.querySelector('.playbutton')
            playbutton.setAttribute('hx-get', 'api/v1/play?code=' + data.Code)
            pausebutton = elem.querySelector('.pausebutton')
            pausebutton.setAttribute('hx-get', 'api/v1/pause?code=' + data.Code)
            stopbutton = elem.querySelector('.stopbutton')
            stopbutton.setAttribute('hx-get', 'api/v1/stop?code=' + data.Code)
            fadeoutstopbutton = elem.querySelector('.fadeoutstopbutton')
            fadeoutstopbutton.setAttribute('hx-get', 'api/v1/fadeoutstop?code=' + data.Code)
            fadeoutstopbutton.setAttribute('hx-include', '#fadeouttime-' + data.Code)
            fadeouttime = elem.querySelector('.fadeouttime')
            fadeouttime.value=data.DefaultFadeTime
            fadeouttime.setAttribute('id', 'fadeouttime-' + data.Code)
            
            container.append(elem)
            elem = document.getElementById('playback-'+data.Code)
            htmx.process(elem)
        }
        elem = document.getElementById('playback-'+data.Code)
        elem.querySelector('.filenamedisplay').innerHTML = data.Name
        progressbar = elem.querySelector('.progressbar')
        progressbar.setAttribute('max',data.Seek.LenProgres)
        progressbar.setAttribute('value',data.Seek.PosProgres)
        progressbar.classList.add('w-100')
        elem.querySelector('.posdisplay').innerHTML = data.Seek.PosDisplay
        elem.querySelector('.remainingdisplay').innerHTML = data.Seek.RemainingDisplay
        elem.querySelector('.volume').innerHTML = data.Volume + '%'
        volslider = elem.querySelector('.volumeslider')
        volslider.value = data.Volume
        
        
        playbutton = elem.querySelector('.playbutton')
        pausebutton = elem.querySelector('.pausebutton')
        if (data.Paused){
            playbutton.style.display = 'block'
            pausebutton.style.display = 'none'
        }
        else{
            playbutton.style.display = 'none'
            pausebutton.style.display = 'block'
        }

        
    });
    existingIds.forEach(v => {
        document.getElementById('playback-' + v).remove()
    })
}

function pingserver(ws){
    //console.log('ping')
    ws.send('ok')
}
function startWebsocket() {
    var refreshIntervalId;
    var wsplaystate = new WebSocket("ws://" + window.location.host + "/ws");
    wsplaystate.onopen = function(e) {
        wsplaystate.send('gimmegimme')
        console.log("Connection open...");
        container = document.getElementById('playing')
        container.innerHTML = ''
        document.getElementById('connectionstatus').classList.remove('bg-dark')
        document.getElementById('connectionstatus').classList.remove('bg-danger')
        document.getElementById('connectionstatus').classList.add('bg-success')
        refreshIntervalId = setInterval( function() {pingserver(wsplaystate)}, 250);
    };
    wsplaystate.onmessage = function(e){
        conectionstatus = document.getElementById('connectionstatus')
        if( ! conectionstatus.classList.contains('bg-danger')){
            if (counter % 6 < 3) {
                conectionstatus.classList.remove('bg-dark')
                conectionstatus.classList.add('bg-success')
            }
            else {
                conectionstatus.classList.remove('bg-success')
                conectionstatus.classList.add('bg-dark')
            }
            if (counter >= 6) {
                counter = 0;
            }        
        }
        counter++
        createPlayback(e.data)
        createLibrary(e.data)
        
            
    }

    wsplaystate.onclose = function(){
        document.getElementById('connectionstatus').classList.remove('bg-dark')
        document.getElementById('connectionstatus').classList.remove('bg-success')
        document.getElementById('connectionstatus').classList.add('bg-danger')
        // connection closed, discard old websocket and create a new one in 5s
        wsplaystate = null
        console.log("Connection closed.");
        clearInterval(refreshIntervalId);
        setTimeout(startWebsocket, 2000)
    }

    window.addEventListener("unload", function () {
            if(wsplaystate.readyState == WebSocket.OPEN){
            wsplaystate.close();
        }
            if(wslibrary.readyState == WebSocket.OPEN){
                wslibrary.close();
        }
    });
}

htmx.on('#uploadform', 'htmx:afterRequest', function(evt){
    clearLibrary()
    htmx.ajax('GET', '/api/v1/reload', '#myDiv')
    document.getElementById('uploadform').reset();
    document.getElementById('uplaodfilename').innerHTML = 'Select file to upload'
});


startWebsocket();