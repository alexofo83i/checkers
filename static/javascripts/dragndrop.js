var whodo = "checker-white";
var whodonot = "checker-black";


function makeNext(){
    $("#donextbutton").attr('disabled',true);
    updateGameState(true);
}

function makeBack(){
    $("#dobackbutton").attr('disabled',true);
    renderGameState(gameStatePrev, true);
    updateGameState(true);
}


//First of all lets make the plyer square element draggable
function makeCheckersDragNDrop(){
    $(".checker").draggable( getdraggable() );
    $(".square").droppable({
        drop : function( event, ui){
            // alert( "drop " + ui.draggable[0].id + " to " + event.target.id);
            checker = ui.draggable[0];
            tdtgt = event.target;
            if( $(checker).hasClass(whodonot) ) return;
            // lose your checker 
            if( $("td#" + event.target.id + " div.checker-green").length == 1 && $(".checker-red").length >= 1 ) {
                alert("Возьмем Вашу шашку " + checker.id + " за фука!");
                //$(".checker-gray").removeClass("checker-gray").addClass(whodonot);
                tdtgt.kickme = null;
                doItWithGameState(checker.id, checker.parentNode.id, null);
                checker.remove();
                switchwhodo();
                makeNext();
            }else
            // just make step
            if( $("td#" + event.target.id + " div.checker-green").length == 1){
                 chs = $("." + whodo);
                 did = false;
                 chs.each(function(i){
                    who = chs[i]; // this is because each returns i
                    iswho = (who.id != null && whoIsWho(who))? true: false;
                    $(".checker-green").remove();
                    $(".checker-red").remove();
                    isnotchecker = who != checker;
                    if( iswho && !did){
                        alert("Возьмем Вашу шашку " + who.id + " за фука!");
                        doItWithGameState(who.id, who.parentNode.id,null);
                        who.remove(); 
                        did = true;
                    }
                });
                doItWithGameState(checker.id, checker.parentNode.id, event.target.id);
                event.target.append(checker);
                switchwhodo();
                makeNext();
            }else
            // kick opponent's checker  
            if($("td#" + event.target.id + " div.checker-red").length == 1 ){
                doItWithGameState(checker.id, checker.parentNode.id, event.target.id);
                doItWithGameState(tdtgt.kickme, tdtgt.kickme.parentNode.id, null);
                event.target.append(checker);
                tdtgt.kickme.remove();
                //if( !whoIsWho(checker) )
                can = canMakeKickAfterKick(checker, tdtgt.kickme);
                if( !can )
                {
                    //checker.kickme = null;
                    switchwhodo();
                    makeNext();
                }
                tdtgt.kickme = null;
            }else
            // nothing to do 
            {
                // $(".checker-gray").removeClass("checker-gray").addClass(whodonot);
                tdtgt.kickme = null;
            }
        }
    });
}



function switchwhodo(){
    var tmpwhodo = whodo;
    whodo = whodonot;
    whodonot = tmpwhodo;
}


function getdraggable(){
    return {
        addClasses: false,
        cancel: "button", // these elements won't initiate dragging
        revert: "invalid", // when not dropped, the item will revert back to its initial position
        containment: "document",
        helper: "clone",
        cursor: "move",
        stack : ".checker",
        start : function ( event ){
            var checker = event.target;
            whoIsWho(checker);
            //askAIwhoiswho(checker);
        },
        drag : function ( event ){
            // console.log( "drag: " + event.target.id );
        },
        stop : function(event){
            clearSteps();
        }
    };
}

function clearSteps(){
    $(".checker-green").remove();
    $(".checker-red").remove();
}

function makeKick(tagTdAlien, tagTdSrc, tagChecker ){
    if( $(tagTdAlien.children[0]).hasClass("checker-black") && $(tagChecker).hasClass("checker-white")
     || $(tagTdAlien.children[0]).hasClass("checker-white") && $(tagChecker).hasClass("checker-black") ){
        tddstid = tagTdAlien.id;
        tdsrcid = tagTdSrc.id;
        shiftY = getFieldShiftY(tddstid,tdsrcid);
        shiftX = getFieldShiftX(tddstid,tdsrcid);
        nextField = getNextField(tddstid, shiftY, shiftX);
        //showGreenRedCheckers(nextField, tdsrc, checker, "checker-red")
        if( nextField != null){
            tagTdKick = $("#" + nextField)[0]
            if ( tagTdKick != null ) {
                if( tagTdKick.children.length == 0 ){
                    makeStep(tagTdKick, tagChecker, "checker-red");
                    tagTdKick.kickme = tagTdAlien.children[0]; // get gray checker
                    return true;
                }
            }
        }
     }
     return false;
}

function makeStep(tagTdTo, tagChecker, chclass){
    var chclone = $(tagChecker).clone();
    chclone
    .removeClass("checker-black")
    .removeClass("checker-white")
    .addClass(chclass);  
    chclone.attr("id","tmp"+tagChecker.id);
    $(tagTdTo).append(chclone);
    return false;
}

function canMakeKickAfterKick(tagChecker, tagChKick){
    res = false;
    $(gameState.next)
    .each(function(i){
        if( gameState.next[i].ch  == tagChecker.id){
            $(gameState.next[i].kicks)
            .each(function(j){
                ch2tdidtrgStr = gameState.next[i].kicks[j]
                var ch2tdidtrgArray = ch2tdidtrgStr.split(",")
                if ( ch2tdidtrgArray.length > 0 ) {
                    ch2tdidtrgArrayArray = ch2tdidtrgArray[0].split("-")
                    ch = ch2tdidtrgArrayArray[0]
                    if ( tagChKick.id == ch ){
                        gameState.next[i].kicks[j] = ch2tdidtrgStr.substring(2+ ch2tdidtrgStr.indexOf(tagChecker.id)+ch2tdidtrgStr.indexOf(",")); // w1-1b,
                        res |= gameState.next[i].kicks[j].length != 0;
                    }
                };
            });
        }
    });
    return res;
}

// function askAIwhoiswho(tagChecker){
//     clearSteps();
//     $(gameState.next)
//     .each(function(i){
//         if( gameState.next[i].ch  == tagChecker.id){
//             $(gameState.next[i].steps)
//             .each(function(j){
//                 tdidtrg = gameState.next[i].steps[j]
//                 td = $("#" + tdidtrg)[0]
//                 makeStep(td, tagChecker, "checker-green");
//             });
//             $(gameState.next[i].kicks)
//             .each(function(j){
//                 ch2tdidtrgStr = gameState.next[i].kicks[j]
//                 var ch2tdidtrgArray = ch2tdidtrgStr.split(",")
//                 if ( ch2tdidtrgArray.length > 0 ) {
//                     ch2tdidtrgArrayArray = ch2tdidtrgArray[0].split("-")
//                     ch = ch2tdidtrgArrayArray[0]
//                     tdidtrg = ch2tdidtrgArrayArray[1]
//                     tagTdKick = $("#" + tdidtrg)[0]
//                     tagChKick = $("#" + ch)[0]
//                     makeStep(tagTdKick, tagChecker, "checker-red");
                    
//                     tagTdKick.kickme = tagChKick; // get gray checker
//                 };
//             });
//         }
//     });
// }

function whoIsWho(checker){
    if( ! $(checker).hasClass(whodo) ){
        return;
    }
    var tdsrc = checker.parentNode;
    var tdsrcid = tdsrc.id;
    fields = [
        getNextField(tdsrcid, -1, -1),
        getNextField(tdsrcid, -1, +1),
        getNextField(tdsrcid, +1, +1),
        getNextField(tdsrcid, +1, -1)
    ];
    return fields
    .filter(function ( tdidtrg ){
        return testYourMight(tdidtrg, tdsrc, checker, "checker-green");
    } )
    .length != 0;
}

function canMakeStep(tdidsrc, tdidtrg) {
    var ysrc = tdidsrc[0]
    var ytrg = tdidtrg[0]
    if( whodo == "checker-white"){
        return ytrg - ysrc > 0
    }else{
        return ytrg - ysrc < 0
    }
}

function testYourMight(idTdAlien, tagTdFrom, TagChecker, chclass){
    if( idTdAlien != null){
        tagTdAlien = $("#" + idTdAlien)[0]
        if ( tagTdAlien != null ) {
            if( tagTdAlien.children.length == 0 ){
                if( canMakeStep(tagTdFrom.id, idTdAlien)){
                    return makeStep(tagTdAlien, TagChecker, chclass);
                }
            }else {
                return makeKick(tagTdAlien, tagTdFrom, TagChecker);
            }
        };
    }
    return false;
}






function getFieldShiftX( fielddstid, fieldsrcid ){
    var dstx = fielddstid[1].charCodeAt();
    var srcx = fieldsrcid[1].charCodeAt();
    return +dstx - +srcx
}

function getFieldShiftY( fielddstid, fieldsrcid ){
    var dsty = fielddstid[0];
    var srcy = fieldsrcid[0];
    return +dsty - +srcy
}



function getNextField( fieldid, shiftY, shiftX ){
    var y = fieldid[0];
    var x = fieldid[1];
    var newY =  +y + +shiftY;
    var newX = String.fromCharCode(x.charCodeAt() + shiftX);
    newY = (newY >= 1 && newY <= 8) ? newY: null;
    if( newY == null ) return null;
    newX = (newX >= 'a' && newX <= 'h') ? newX: null;
    if( newX == null ) return null;
    return newY + newX;
}