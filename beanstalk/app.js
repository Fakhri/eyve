var express = require('express');
var exphbs = require('express-handlebars');
var app = express();
var scripts = [{ script: '/js/script.js' }];
var liveScripts = [{ script: '/js/live.js' }];
require('dotenv').config();

// Set view engine
app.engine('hbs', exphbs());
app.set('view engine', 'hbs');

// Require multer for file uploading and multers3 to upload directly to s3
var multer = require('multer');
var multerS3 = require('multer-s3');

// Configure aws s3 SDK (update authentication)
var AWS = require('aws-sdk');
AWS.config.update({
  accessKeyId: process.env.AWS_ACCESS_KEY,
  secretAccessKey: process.env.AWS_SECRET_KEY,
  region: process.env.AWS_REGION
});
var s3 = new AWS.S3();

// Unique name of aws s3 bucket created
const myBucket = process.env.AWS_BUCKET;

// Multer upload (Use multer-s3 to save directly to AWS instead of locally)
var upload = multer({
  storage: multerS3({
    s3: s3,
    bucket: myBucket,
    // Set public read permissions
    // acl: 'public-read',
    // Auto detect contet type
    // contentType: multerS3.AUTO_CONTENT_TYPE, 
    // Set key/ filename as original uploaded name
    key: function (req, file, cb) {
      cb(null, file.originalname.replace(/\s+/g, '-'))
    }
  })
})

// Express routes
app.get('/', (req, res, next) => {
  res.render('index', {scripts: scripts});
});

app.get('/live', (req, res, next) => {
  res.render('live', {scripts: liveScripts});
});

// Upload single file endpoint (calls on upload middleware above)
// upload.single('name') is the key that the file accepts
app.post('/single', upload.single('video'), (req, res, next) => {
  res.redirect('/album');
});

// Upload multiple max 3
app.post('/multiple', upload.array('video', 3), (req, res, next) => {
  res.send(`Succesfully uploaded files: ${req.files.length}`);
});

// View all files 
app.get('/album', (req, res, next) => {
  // // Traditional CAllBACK method
  // s3.listObjects({Bucket: myBucket}, function(err, data) {
  //   if (err) { console.log(err) }
  //   // Retrieve all files filenames and create url array
  //   const baseURL = `https://s3.amazonaws.com/${myBucket}/`;      
  //   let urlArr = data.Contents.map(e => baseURL + e.Key);
  //   res.render('album', { data: urlArr});
  // })

  // USING PROMISES, call on the promise method
  s3.listObjects({Bucket: myBucket}).promise()
    .then(data => {
      const baseURL = `https://s3.amazonaws.com/${myBucket}/`;      
      let urlArr = data.Contents.map(e => baseURL + e.Key);
      res.render('album', { data: urlArr});
    })
    .catch(err => console.log(err));
  
});

// Return file object
app.get('/view/:filename', (req, res, next) => {
  var params = { Bucket: myBucket, Key: req.params.filename };
  s3.getObject(params, function(err, data){
    if (err) { 
      return next() 
    } else {
      // Convert file to base65 rile
      var img = new Buffer(data.Body, 'base64');
      res.contentType(data.ContentType);
      res.status(200).send(img);
    } 
  });
});

// Return file url
app.get('/url/:filename', (req, res, next) => {
  var params = { Bucket: myBucket, Key: req.params.filename };
  s3.getSignedUrl('getObject', params, function(err, url){
    if (err) { console.log(err) }
    res.send(url)
  })
});

app.use(express.static(__dirname + '/public'));

app.use((req,res) => {
  res.status(404).send('Error 404');
});

app.listen(process.env.PORT, () => {
  console.log('Listening on localhost:' + process.env.PORT);
})
