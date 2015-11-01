var UUID = (function() {
    var self = {};
    var lut = []; for (var i=0; i<256; i++) { lut[i] = (i<16?'0':'')+(i).toString(16); }
    self.generate = function() {
    var d0 = Math.random()*0xffffffff|0;
    var d1 = Math.random()*0xffffffff|0;
    var d2 = Math.random()*0xffffffff|0;
    var d3 = Math.random()*0xffffffff|0;
        return lut[d0&0xff]+lut[d0>>8&0xff]+lut[d0>>16&0xff]+lut[d0>>24&0xff]+'-'+
               lut[d1&0xff]+lut[d1>>8&0xff]+'-'+lut[d1>>16&0x0f|0x40]+lut[d1>>24&0xff]+'-'+
               lut[d2&0x3f|0x80]+lut[d2>>8&0xff]+'-'+lut[d2>>16&0xff]+lut[d2>>24&0xff]+
               lut[d3&0xff]+lut[d3>>8&0xff]+lut[d3>>16&0xff]+lut[d3>>24&0xff];
    }
    return self;
})();

function Prospects()
{
    this.setupGeolocation();
    this.setupForm(document.getElementById("mainForm"));
    this.setupForm(document.getElementById("feedbackForm"));
};

Prospects.prototype.setupGeolocation = function()
{
    var prospectsThis = this;

    if(navigator.geolocation) {
        console.log("Geolocation: " + navigator.geolocation)

        navigator.geolocation.getCurrentPosition(
            function(pos) {
                prospectsThis.crd = pos.coords;

                console.log('Latitude: ' + prospectsThis.crd.latitude);
                console.log('Longitude: ' + prospectsThis.crd.longitude);
                console.log('More or less ' + prospectsThis.crd.accuracy + ' meters.');

            },
            function(err) {
                console.warn('ERROR(' + err.code + '): ' + err.message);
            },
            {enableHighAccuracy: true, timeout: 5000, maximumAge: 0}
        );
    }
};

Prospects.prototype.getUUID = function()
{
    var storage = window.localStorage;
    var uuid = storage.getItem('uuid');

    if(null == uuid) {
        uuid = UUID.generate();
        storage.setItem('uuid', uuid);
    }

    return uuid;
};

Prospects.prototype.getParameterFromInput = function(inputElement)
{
    return inputElement.name + "=" + encodeURIComponent(inputElement.value);
};

Prospects.prototype.setupForm = function(formElement)
{
    var inputs = formElement.getElementsByTagName("input");

    var appnameText = null;
    var firstnameText = null;
    var middlenameText = null;
    var lastnameText = null;
    var emailText = null;
    var phonenumberText = null;
    var clearButton = null;

    var pinterestBox = null;
    var facebookBox = null;
    var instagramBox = null;
    var twitterBox = null;
    var googleBox = null;
    var youtubeBox = null;
    var feedbackBox = null;

    var femaleRadio = null;
    var maleRadio = null;

    var dobDate = null;

    for(var iter = 0; iter < inputs.length; ++iter)
    {
        if(inputs[iter].name == "appname")
            appnameText = inputs[iter];
        else if(inputs[iter].name == "firstname")
            firstnameText = inputs[iter];
        else if(inputs[iter].name == "middlename")
            middlenameText = inputs[iter];
        else if(inputs[iter].name == "lastname")
            lastnameText = inputs[iter];
        else if(inputs[iter].name == "email")
            emailText = inputs[iter];
        else if(inputs[iter].name == "phonenumber")
            phonenumberText = inputs[iter];
        else if(inputs[iter].name == "pinterest")
            pinterestBox = inputs[iter];
        else if(inputs[iter].name == "facebook")
            facebookBox = inputs[iter];
        else if(inputs[iter].name == "instagram")
            instagramBox = inputs[iter];
        else if(inputs[iter].name == "twitter")
            twitterBox = inputs[iter];
        else if(inputs[iter].name == "google")
            googleBox = inputs[iter];
        else if(inputs[iter].name == "youtube")
            youtubeBox = inputs[iter];
        else if(inputs[iter].name == "gender" && inputs[iter].value == "female")
            femaleRadio = inputs[iter];
        else if(inputs[iter].name == "gender" && inputs[iter].value == "male")
            maleRadio = inputs[iter];
        else if(inputs[iter].name == "dob")
            dobDate = inputs[iter];
        else if(inputs[iter].value.toLowerCase() == "clear")
            clearButton = inputs[iter];
    };

    var feedbackTextArea = formElement.getElementsByTagName("textarea");
    if(feedbackTextArea.length > 0) {
        feedbackBox = feedbackTextArea[0];
    }

    var contestSelect = formElement.getElementsByTagName("select")[0];

    var prospectsThis = this;

    formElement.addEventListener("submit", function(e)
    {
        var xmlHttp = new XMLHttpRequest();

        xmlHttp.onload = function(lev)
        {
            console.log(xmlHttp.responseText);
            console.log(xmlHttp.status);
            jsonResponse = JSON.parse(xmlHttp.responseText);
            console.log(jsonResponse);
            formElement.reset();
        };

        xmlHttp.onerror = function(eev)
        {
            console.log(xmlHttp.responseText);
            console.log(xmlHttp.status);
            alert("Server error");
        };

        var queryStr = "";
        
        queryStr += "leadid=" + encodeURIComponent(prospectsThis.getUUID()) + "&";

        if(null != appnameText && appnameText.value.length > 0)
            queryStr += prospectsThis.getParameterFromInput(appnameText) + "&";

        if(null != firstnameText && firstnameText.value.length > 0)
            queryStr += prospectsThis.getParameterFromInput(firstnameText) + "&";

        if(null != middlenameText && middlenameText.value.length > 0)
            queryStr += prospectsThis.getParameterFromInput(middlenameText) + "&";

        if(null != lastnameText && lastnameText.value.length > 0)
            queryStr += prospectsThis.getParameterFromInput(lastnameText) + "&";

        if(null != emailText && emailText.value.length > 0)
            queryStr += prospectsThis.getParameterFromInput(emailText) + "&";

        if(null != phonenumberText && phonenumberText.value.length > 0)
            queryStr += prospectsThis.getParameterFromInput(phonenumberText) + "&";

        if(null != pinterestBox && pinterestBox.checked)
            queryStr += prospectsThis.getParameterFromInput(pinterestBox) + "&";

        if(null != facebookBox && facebookBox.checked)
            queryStr += prospectsThis.getParameterFromInput(facebookBox) + "&";

        if(null != instagramBox && instagramBox.checked)
            queryStr += prospectsThis.getParameterFromInput(instagramBox) + "&";

        if(null != twitterBox && twitterBox.checked)
            queryStr += prospectsThis.getParameterFromInput(twitterBox) + "&";

        if(null != googleBox && googleBox.checked)
            queryStr += prospectsThis.getParameterFromInput(googleBox) + "&";

        if(null != youtubeBox && youtubeBox.checked)
            queryStr += prospectsThis.getParameterFromInput(youtubeBox) + "&";

        if(null != femaleRadio && femaleRadio.checked)
            queryStr += prospectsThis.getParameterFromInput(femaleRadio) + "&";
        else if(null != maleRadio && maleRadio.checked)
            queryStr += prospectsThis.getParameterFromInput(maleRadio) + "&";

        if(null != feedbackBox && feedbackBox.value.length > 0)
            queryStr += prospectsThis.getParameterFromInput(feedbackBox) + "&";

        if(null != dobDate && dobDate.value) {
            var dob = new Date(dobDate.value);
            queryStr += dobDate.name + "=" + encodeURIComponent(dob.toISOString()) + "&";
        }

        if(null != contestSelect && contestSelect.value.length > 0) {
            console.log(contestSelect.name);
            console.log(contestSelect.value);
            var contest = { "contest" : {"eligible" : true,
                                         "item" : contestSelect.value
                                        }
                          };

            console.log(contest);

            queryStr += "miscellaneous=" + JSON.stringify(contest) + "&";
        }

        if(document.referrer) {
            console.log("Referrer: " + document.referrer);
            queryStr += "pagereferrer=" + encodeURIComponent(document.referrer) + "&";
        } else {
            console.log("Page was not referred");
        }

        if(navigator.language) {
            console.log("Language: " + navigator.language)
            queryStr += "language=" + encodeURIComponent(navigator.language) + "&";
        }

        if(navigator.geolocation) {
            if(null != prospectsThis.crd) {
                queryStr += "latitude=" + encodeURIComponent(prospectsThis.crd.latitude) + "&";
                queryStr += "longitude=" + encodeURIComponent(prospectsThis.crd.longitude) + "&";
            }
        }

        if(navigator.platform) {
            console.log("Platform: " + navigator.platform)
        }

        xmlHttp.open(formElement.method, formElement.action, true);
        xmlHttp.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
        xmlHttp.withCredentials = true;
        xmlHttp.send(queryStr);
        console.log(queryStr);

        e.preventDefault();
    });

    clearButton.addEventListener("click", function(e)
    {
        formElement.reset();
    });
};
