/**
 * var prospect = new Prospect();
 *
 * prospect.setUrl("https://host:port/prospects");
 * prospect.setAppname("bronxwood");
 * prospect.setEmail("raul.ferris@gmail.com");
 * prospect.setPhoneNumber("212-555-1212");
 *
 * prospect.ready(); //Returns if ready to save or not
 * prospect.save(); //Synchronous.  Returns object with response data.
 *
 * prospect.save( //Asynchronous
 *      function(response, status, pros) {}, //Success
 *      function(response, status, pros) {}  //Error
 * );
 */

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

function NotEmpty(value)
{
    return typeof value === 'string' && value.length > 0;
};

function IsBoolean(value)
{
    return typeof value === 'boolean';
};

function IsBooleanTrue(value)
{
    return IsBoolean(value) && value;
};

function getParameterFromNvp(name, value)
{
    return name + "=" + encodeURIComponent(value) + "&";
};

function Prospect()
{
    this.uuid = this.getUUID();
    this.adhocFields = {};
    this.adhocHeaders = {};
};

Prospect.prototype.getUUID = function()
{
    var storage = window.localStorage;
    var uuid = storage.getItem('uuid');

    if(null == uuid) {
        uuid = UUID.generate();
        storage.setItem('uuid', uuid);
    }

    return uuid;
};

Prospect.prototype.setUrl = function(url) {
    this.url = url;
};

Prospect.prototype.getUrl = function() {
    return this.url;
}

Prospect.prototype.setAppName = function(appName) {
    this.appName = appName;
};

Prospect.prototype.getAppName = function() {
    return this.appName;
};

Prospect.prototype.setEmail = function(email) {
    this.email = email;
};

Prospect.prototype.getEmail = function() {
    return this.email;
};

Prospect.prototype.setPinterest = function(pinterest) {
    this.pinterest = pinterest;
};

Prospect.prototype.getPinterest = function() {
    return this.pinterest;
};

Prospect.prototype.setFacebook = function(facebook) {
    this.facebook = facebook;
};

Prospect.prototype.getFacebook = function() {
    return this.facebook;
};

Prospect.prototype.setInstagram = function(instagram) {
    this.instagram = instagram;
};

Prospect.prototype.getInstagram = function() {
    return this.instagram;
};

Prospect.prototype.setTwitter = function(twitter) {
    this.twitter = twitter;
};

Prospect.prototype.getTwitter = function() {
    return this.twitter;
};

Prospect.prototype.setGoogle = function(google) {
    this.google = google;
};

Prospect.prototype.getGoogle = function() {
    return this.google;
};

Prospect.prototype.setYoutube = function(youtube) {
    this.youtube = youtube;
};

Prospect.prototype.getYoutube = function() {
    return this.youtube;
};

Prospect.prototype.setExtended = function(extended) {
    this.extended = extended;
};

Prospect.prototype.getExtended = function() {
    return this.extended;
};

Prospect.prototype.setFeedback = function(feedback) {
    this.feedback = feedback;
};

Prospect.prototype.getFeedback = function() {
    return this.feedback;
};

Prospect.prototype.setPageReferrer = function(pageReferrer) {
    this.pageReferrer = pageReferrer;
};

Prospect.prototype.getPageReferrer = function() {
    return this.pageReferrer;
};

Prospect.prototype.setFirstName = function(firstName) {
    this.firstName = firstName;
};

Prospect.prototype.getFirstName = function() {
    return this.firstName;
};

Prospect.prototype.setLastName = function(lastName) {
    this.lastName = lastName;
};

Prospect.prototype.getLastName = function() {
    return this.lastName;
};

Prospect.prototype.setPhoneNumber = function(phoneNumber) {
    this.phoneNumber = phoneNumber;
};

Prospect.prototype.getPhoneNumber = function() {
    return this.phoneNumber;
};

Prospect.prototype.setDateOfBirth = function(dateOfBirth) {
    this.dateOfBirth = dateOfBirth;
};

Prospect.prototype.getDateOfBirth = function() {
    return this.dateOfBirth;
};

Prospect.prototype.setGender = function(gender) {
    this.gender = gender;
};

Prospect.prototype.getGender = function() {
    return this.gender;
};

Prospect.prototype.setZipCode = function(zipCode) {
    this.zipCode = zipCode;
};

Prospect.prototype.getZipCode = function() {
    return this.zipCode;
};

Prospect.prototype.setLanguage = function(language) {
    this.language = language;
};

Prospect.prototype.getLanguage = function() {
    return this.language;
};

Prospect.prototype.setLatitude = function(latitude)
{
    this.latitude = latitude;
};

Prospect.prototype.getLatitude = function()
{
    return this.latitude;
};

Prospect.prototype.setLongitude = function(longitude)
{
    this.longitude = longitude;
};

Prospect.prototype.getLongitude = function()
{
    return this.longitude;
};

Prospect.prototype.setMiscellaneous = function(miscellaneous) {
    this.miscellaneous = miscellaneous;
};

Prospect.prototype.getMiscellaneous = function() {
    return this.miscellaneous;
};

Prospect.prototype.addAdhocField = function(name, value) {
    this.adhocFields[name] = value;
};

Prospect.prototype.getAdhocFields = function() {
    return this.adhocFields;
};

Prospect.prototype.addAdhocHeader = function(name, value) {
    this.adhocHeaders[name] = value;
};

Prospect.prototype.getAdhocHeaders = function() {
    return this.adhocHeaders;
};

Prospect.prototype.ready = function() {
    return NotEmpty(this.url) && NotEmpty(this.uuid) && NotEmpty(this.appName)
            && (NotEmpty(this.email)
             || NotEmpty(this.phoneNumber)
             || IsBooleanTrue(this.pinterest)
             || IsBooleanTrue(this.facebook)
             || IsBooleanTrue(this.twitter)
             || IsBooleanTrue(this.instagram)
             || IsBooleanTrue(this.google)
             || IsBooleanTrue(this.youtube)
             || IsBooleanTrue(this.extended)
             || NotEmpty(this.feedback));
};

Prospect.prototype.save = function(successFunc, errorFunc) {
    if(!this.ready()) {
        throw new Error("Prospect has missing required fields");
    }

    var xmlHttp = new XMLHttpRequest();
    var async = (null != successFunc || null != errorFunc);
    var that = this;

    xmlHttp.onload = function(e)
    {
        if(null != successFunc) {
            successFunc(JSON.parse(xmlHttp.responseText), xmlHttp.status, that);
        }
    };

    xmlHttp.onerror = function(e)
    {
        if(null != errorFunc) {
            if(0 != xmlHttp.status) {
                errorFunc(JSON.parse(xmlHttp.responseText), xmlHttp.status, that);
            } else {
                //Titanium handles connection down errors here.  Browsers typically throw an exception on send
                var error = { "code":503,
                              "code_message":"Service unavailable",
                              "message":e.error
                            };

                errorFunc(error, error.code, that);
            }
        }
    };

    var queryStr = getParameterFromNvp("leadid", this.uuid);
    queryStr += getParameterFromNvp("appname", this.appName);

    if(NotEmpty(this.email))
        queryStr += getParameterFromNvp("email", this.email);

    if(IsBoolean(this.pinterest))
        queryStr += getParameterFromNvp("pinterest", this.pinterest);

    if(IsBoolean(this.facebook))
        queryStr += getParameterFromNvp("facebook", this.facebook);

    if(IsBoolean(this.instagram))
        queryStr += getParameterFromNvp("instagram", this.instagram);

    if(IsBoolean(this.twitter))
        queryStr += getParameterFromNvp("twitter", this.twitter);

    if(IsBoolean(this.google))
        queryStr += getParameterFromNvp("google", this.google);

    if(IsBoolean(this.youtube))
        queryStr += getParameterFromNvp("youtube", this.youtube);

    if(IsBoolean(this.extended))
        queryStr += getParameterFromNvp("extended", this.extended);

    if(NotEmpty(this.firstName))
        queryStr += getParameterFromNvp("firstname", this.firstName);

    if(NotEmpty(this.lastName))
        queryStr += getParameterFromNvp("lastname", this.lastName);

    if(NotEmpty(this.feedback))
        queryStr += getParameterFromNvp("feedback", this.feedback);

    if(NotEmpty(this.phoneNumber))
        queryStr += getParameterFromNvp("phonenumber", this.phoneNumber);

    if(NotEmpty(this.dateOfBirth)) {
        var dob = new Date(this.dateOfBirth).toISOString();
        queryStr += getParameterFromNvp("dob", dob);
    }

    if(NotEmpty(this.gender))
       queryStr += getParameterFromNvp("gender", this.gender);

    if(NotEmpty(this.zipcode))
       queryStr += getParameterFromNvp("zipcode", this.zipcode);

    if(NotEmpty(this.language))
       queryStr += getParameterFromNvp("language", this.language);
    else if(navigator.language)
       queryStr += getParameterFromNvp("language", navigator.language);

    if(NotEmpty(this.pageReferrer))
        queryStr += getParameterFromNvp("pagereferrer", this.pageReferrer);
    else if(document.referrer)
        queryStr += getParameterFromNvp("pagereferrer", document.referrer);

    if(NotEmpty(this.latitude))
        queryStr += getParameterFromNvp("latitude", this.latitude);

    if(NotEmpty(this.longitude))
        queryStr += getParameterFromNvp("longitude", this.longitude);

    if(NotEmpty(this.miscellaneous))
        queryStr += getParameterFromNvp("miscellaneous", this.miscellaneous);

    for(var key in this.adhocFields) {
        if(this.adhocFields.hasOwnProperty(key)) {
            queryStr += getParameterFromNvp(key, this.adhocFields[key]);
        }
    }

    try
    {
        xmlHttp.open("POST", this.url, async);
        xmlHttp.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');

        for(var key in this.adhocHeaders) {
            if(this.adhocHeaders.hasOwnProperty(key)) {
                xmlHttp.setRequestHeader(key, this.adhocHeaders[key]);
            }
        }

        xmlHttp.withCredentials = true;
        xmlHttp.send(queryStr);

        if(!async) {
            if(0 != xmlHttp.status) {
                return JSON.parse(xmlHttp.responseText);
            } else {
                var error = { "code":503,
                              "code_message":"Service unavailable",
                              "message":exp.name
                            };

                return error;
            }
        }
    }
    catch(exp)
    {
        //Browsers typically handle connection refused errors by throwing an exception on send
        var error = { "code":503,
                      "code_message":"Service unavailable",
                      "message":exp.name
                    };

        if(!async) {
            return error;
        } else {
            if(null != errorFunc) {
                errorFunc(error, error.code, that);
            }
        }
    }
};
