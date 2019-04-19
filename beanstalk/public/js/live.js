AWS.config.update({
  credentials: {
      accessKeyId: "<YOUR_AWS_ACCESS_KEY_ID>",
      secretAccessKey: "<YOUR_AWS_SECRET_ACCESS_KEY>",
      region: "<YOUR_AWS_REGION>"
  }
});

const BUCKET_NAME = "eyve-image-analysis";
const s3 = new AWS.S3();
let isRecording = false;
const s3Key = `videofile${new Date().toISOString()}`.replace(/[\W_]+/g, "") + '.webn';
const params = {
  Bucket: BUCKET_NAME,
  Key: s3Key
};

let videoStream;
// We want to see what camera is recording so attach the stream to video element.
navigator.mediaDevices
  .getUserMedia({
    audio: true,
    video: { width: 426, height: 240 }
  })
  .then(stream => {
    console.log("Successfully received user media.");

    const $mirrorVideo = document.querySelector("video#mirror");
    $mirrorVideo.srcObject = stream;

    // Saving the stream to create the MediaRecorder later.
    videoStream = stream;
  })
  .catch(error => console.error("navigator.getUserMedia error: ", error));

let mediaRecorder;

const $startButton = document.querySelector("button#start");
// $startButton.onclick = start; 
$startButton.onclick = startPhotoStream; 

const $pictButton = document.getElementById("screenshot-button");
$pictButton.onclick = recordImages; 

function startPhotoStream() {
  setInterval(function () {
    recordImages();
  }, 10000);
}

function start() {
  $startButton.setAttribute('disabled', true);
  isRecording = true;
  
  // Getting the MediaRecorder instance.
  // I took the snippet from here: https://github.com/webrtc/samples/blob/gh-pages/src/content/getusermedia/record/js/main.js
  let options = { mimeType: "video/webm;codecs=vp9" };
  if (!MediaRecorder.isTypeSupported(options.mimeType)) {
    console.log(options.mimeType + " is not Supported");
    options = { mimeType: "video/webm;codecs=vp8" };
    if (!MediaRecorder.isTypeSupported(options.mimeType)) {
      console.log(options.mimeType + " is not Supported");
      options = { mimeType: "video/webm" };
      if (!MediaRecorder.isTypeSupported(options.mimeType)) {
        console.log(options.mimeType + " is not Supported");
        options = { mimeType: "" };
      }
    }
  }

  try {
    mediaRecorder = new MediaRecorder(videoStream, options);
  } catch (e) {
    console.error("Exception while creating MediaRecorder: " + e);
    return;
  }

  //Generate the file name to upload. For the simplicity we're going to use the current date.


  let uploadId;

  // We are going to handle everything as a chain of Observable operators.
  Rx.Observable
    // First create the multipart upload and wait until it's created.
    .fromPromise(s3.createMultipartUpload(params).promise())
    .switchMap(data => {
      // Save the uploadId as we'll need it to complete the multipart upload.
      uploadId = data.UploadId;
      mediaRecorder.start(15000);

      // Then track all 'dataavailable' events. Each event brings a blob (binary data) with a part of video.
      return Rx.Observable.fromEvent(mediaRecorder, "dataavailable");
    })
    // Track the dataavailable event until the 'stop' event is fired.
    // MediaRecorder emits the "stop" when it was stopped AND have emitted all "dataavailable" events.
    // So we are not losing data. See the docs here: https://developer.mozilla.org/en-US/docs/Web/API/MediaRecorder/stop
    .takeUntil(Rx.Observable.fromEvent(mediaRecorder, "stop"))
    .map((event, index) => {
      // Show how much binary data we have recorded.
      const $bytesRecorded = document.querySelector("span#bytesRecorded");
      $bytesRecorded.textContent =
        parseInt($bytesRecorded.textContent) + event.data.size; // Use frameworks in prod. This is just an example.

      // Take the blob and it's number and pass down.
      return { blob: event.data, partNumber: index + 1 };
    })
    // This operator means the following: when you receive a blob - start uploading it.
    // Don't accept any other uploads until you finish uploading: http://reactivex.io/rxjs/class/es6/Observable.js~Observable.html#instance-method-concatMap
    .concatMap(({ blob, partNumber }) => {
      // var newBlob = );
      return (
        s3
          .uploadPart({
            Body: blob,
            Bucket: BUCKET_NAME,
            Key: s3Key,
            PartNumber: partNumber,
            UploadId: uploadId,
            ContentLength: blob.size
          })
          .promise()
          // Save the ETag as we'll need it to complete the multipart upload
          .then(({ ETag }) => {
            // How how much bytes we have uploaded.
            const $bytesUploaded = document.querySelector("span#bytesUploaded");
            $bytesUploaded.textContent =
              parseInt($bytesUploaded.textContent) + blob.size;

            return { ETag, PartNumber: partNumber };
          })
      );
    })
    // Wait until all uploads are completed, then convert the results into an array.
    .toArray()
    // Call the complete multipart upload and pass the part numbers and ETags to it.
    .switchMap(parts => {
      return s3
        .completeMultipartUpload({
          Bucket: BUCKET_NAME,
          Key: s3Key,
          UploadId: uploadId,
          MultipartUpload: {
            Parts: parts
          }
        })
        .promise();
    })
    .subscribe(
      ({ Location }) => {
        // completeMultipartUpload returns the location, so show it.
        const $location = document.querySelector("span#location");
        $location.appendChild(document.createTextNode(Location));
        $location.appendChild(document.createElement("br"));

        console.log("Uploaded successfully.");
      },
      err => {
        console.error(err);

        if (uploadId) {
          // Aborting the Multipart Upload in case of any failure.
          // Not to get charged because of keeping it pending.
          s3
            .abortMultipartUpload({
              Bucket: BUCKET_NAME,
              UploadId: uploadId,
              Key: s3Key
            })
            .promise()
            .then(() => console.log("Multipart upload aborted"))
            .catch(e => console.error(e));
        }
      }
    );
};

const $stopButton = document.querySelector("button#stop");
$stopButton.onclick = () => {
  let startButton = document.querySelector("button#start");
  startButton.disabled = false;
  isRecording = false;
  // After we call .stop() MediaRecorder is going to emit all the data it has via 'dataavailable'.
  // And then finish our stream by emitting 'stop' event.
  mediaRecorder.stop();
};

setInterval(function () {
  if (isRecording) {
    mediaRecorder.stop();
    isRecording = true;
    start();
  }
}, 10000);

function recordImages() {
  let video = document.getElementById('mirror')
  var d = new Date();
  
  var canvas = document.createElement('canvas');
  canvas.width = 480;
  canvas.height = 360;
  var ctx = canvas.getContext('2d');
  ctx.drawImage(video, 0, 0, canvas.width, canvas.height);

  canvas.toBlob(function(blob) {
    var file = new File([blob], "image-" + d.getTime() + ".png", {type: "application/octet-stream"});
    // URL.createObjectURL(file);
    console.log(file);
    s3.upload({
      Body: file,
      Bucket: BUCKET_NAME,
      Key: file.name,
      ACL: 'public-read'
    }, function(err, data) {
      if (err) {
        console.log('failed');
        console.log(err);
      }
      console.log('success');
      console.log(data);
    });
  })
}