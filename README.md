# Eyve
An eye to protect  visually impaired to avoid obstacle, being cheated, and stay safe in public.
This project is awarded as the 3rd place winner at [AWS Hackdays 2019 - Singapore Grand Finale](https://aws.agorize.com/en/challenges/final-hackathon-day-2019/)

### Demo
- https://youtu.be/z-6UOWPExp8
- http://eyve-dev.us-east-1.elasticbeanstalk.com (we temporary shut down this server)
- https://www.facebook.com/Eyve-853664541650117/

### Architecture Diagram
Our plan is to three different approaches in getting the input (for both online and offline solution) and also two development phase of image/video processing that using Rekognition in the first phase then move to AWS SageMaker after the model has been trained and ready for use.

#### Target Implementation
![architecture](https://github.com/fakhri/eyve/blob/master/README_files/Eyve%20-%20Target%20Architecture.png)

### How to Use This Code
To be able to use this code, the AWSs need to be configured as seen in the architecture diagram. This code only contains client app (which need to be deployed to Elastic Beanstalk or any similar technology), Lambda functions, and a notebook for AWS Sagemaker. 
