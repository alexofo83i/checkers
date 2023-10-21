var gameState = null;
var gameStatePrev = null;

function doItWithGameState(checkerid, tdsrcid, tdtgtid){
    if( gameState == null) return;
    $.each( gameState.trs,function( i, tr ){
        $.each( tr.tds, function( j, td){
            if( td.id == tdtgtid ){
                td.ch.id = checkerid;
            }else if( td.id == tdsrcid ){
                td.ch.id = "";
            }
        });     
    });
    gameState.whodo = whodo;
}

function getClassById( checkerid){
    return ( checkerid[0] == "w" )? "checker-white": "checker-black";
}


function updateGameState(letdonext){
   if( gameState == null){
        gameState = { n : $("#chess-board")[0].innerText, trs : [] }
   }else {
        gameStatePrev = gameState
   }
    
   $.ajax({
        type: "POST",
        url: "/game/state/",
        contentType : 'application/json',
        data: JSON.stringify(gameState)
      })
      .done(function( dataRAW ) {
        var data = jQuery.parseJSON(dataRAW);
        renderGameState( data, letdonext );
    })
    .fail(function( msg ) {
        alert( "error: " + msg);
    });
   
}


function renderGameState(data, letdonext ){
    chs = $(".checker")
        chs.each(function( i ){
            if( chs[i].id != "checker-black" && chs[i].id != "checker-white"){
                chs[i].remove();
            }
        });
        $.each( data.trs,function( i, tr ){
            $.each( tr.tds, function( j, td){
                if( td.ch.id != "" ){ 
                    var chdiv = $("#" + getClassById(td.ch.id)).clone();
                    var tdel = $("#" + td.id);
                    tdel.append(chdiv);
                    chdiv.show();
                    chdiv.attr("id", td.ch.id); 
                    chdiv.prop("innerText", td.ch.id);
                }
            });     
        });
        gameStatePrev = gameState
        gameState = data;

        makeCheckersDragNDrop();
        if( letdonext ){
            switchwhodo()
            $("#donextbutton").attr('disabled',false);
            $("#dobackbutton").attr('disabled',false);
        }
}
