function setupForm(formElement)
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
    var latitudeBox = null;
    var longitudeBox = null;
    var geolocationButton = null;

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
        else if(inputs[iter].name == "latitude")
            latitudeBox = inputs[iter];
        else if(inputs[iter].name == "longitude")
            longitudeBox = inputs[iter];
        else if(inputs[iter].value.toLowerCase() == "get location")
            geolocationButton = inputs[iter];
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
        e.preventDefault();

        var prospect = new Prospect();

        prospect.setUrl(formElement.action);
        prospect.setExtended(true);

        if(null != appnameText)
            prospect.setAppName(appnameText.value);

        if(null != firstnameText)
            prospect.setFirstName(firstnameText.value);

        if(null != middlenameText)
            prospect.addAdhocField("middlename", middlenameText.value);

        if(null != lastnameText)
            prospect.setLastName(lastnameText.value);

        if(null != emailText)
            prospect.setEmail(emailText.value);

        if(null != phonenumberText)
            prospect.setPhoneNumber(phonenumberText.value);

        if(null != pinterestBox && pinterestBox.checked)
            prospect.setPinterest(pinterestBox.checked);

        if(null != facebookBox && facebookBox.checked)
            prospect.setFacebook(facebookBox.checked);

        if(null != instagramBox && instagramBox.checked)
            prospect.setInstagram(instagramBox.checked);

        if(null != twitterBox && twitterBox.checked)
            prospect.setTwitter(twitterBox.checked);

        if(null != googleBox && googleBox.checked)
            prospect.setGoogle(googleBox.checked);

        if(null != youtubeBox && youtubeBox.checked)
            prospect.setYoutube(youtubeBox.checked);

        if(null != femaleRadio && femaleRadio.checked)
            prospect.setGender(femaleRadio.value);
        else if(null != maleRadio && maleRadio.checked)
            prospect.setGender(maleRadio.value);

        if(null != feedbackBox)
            prospect.setFeedback(feedbackBox.value);

        if(null != dobDate && dobDate.value)
            prospect.setDateOfBirth(dobDate.value);

        if(null != latitudeBox)
            prospect.setLatitude(latitudeBox.value);

        if(null != longitudeBox)
            prospect.setLongitude(longitudeBox.value);

        if(null != contestSelect && contestSelect.value.length > 0) {
            console.log(contestSelect.name);
            console.log(contestSelect.value);
            var contest = { "contest" : {"eligible" : true,
                                         "item" : contestSelect.value
                                        }
                          };

            console.log(contest);

            prospect.setMiscellaneous(JSON.stringify(contest));
        }

        prospect.save(
            function(response, status, pros)
            {
                console.log(response);
                console.log(status);
                formElement.reset();
            },
            function(response, status, pros)
            {
                console.log(response);
                console.log(status);
                alert("Server error");
            }
        );
    });

    if(null != geolocationButton) {
        geolocationButton.addEventListener("click", function(e)
        {
            if(navigator.geolocation)
            {
                console.log("Geolocation: " + navigator.geolocation)

                navigator.geolocation.getCurrentPosition(
                    function(pos)
                    {
                        latitudeBox.value = pos.coords.latitude;
                        longitudeBox.value = pos.coords.longitude;

                        console.log('Latitude: ' + pos.coords.latitude);
                        console.log('Longitude: ' + pos.coords.longitude);
                        console.log('More or less ' + pos.coords.accuracy + ' meters.');
                    },
                    function(err)
                    {
                        console.warn('ERROR(' + err.code + '): ' + err.message);
                    },
                    {enableHighAccuracy: true, timeout: 5000, maximumAge: 0}
                );
            }
        });
    }

    clearButton.addEventListener("click", function(e)
    {
        formElement.reset();
    });
};

setupForm(document.getElementById("mainForm"));
setupForm(document.getElementById("feedbackForm"));
