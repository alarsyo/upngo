<template>
  <div>
    <h1>Test :)</h1>
    <v-btn @click="$refs.fileInput.click()">
      Test
      <input type="file"
             multiple
             style="display:none"
             ref="fileInput"
             @change='uploadFile'/>
    </v-btn>
  </div>
</template>

<script>
import tus from 'tus-js-client';

export default {
  components: {
  },
  methods: {
    uploadFile: (e) => {
      // Get the selected file from the input element
      const file = e.target.files[0];

      // Create a new tus upload
      const upload = new tus.Upload(file, {
        endpoint: 'https://up.alarsyo.com/files/',
        retryDelays: [0, 3000, 5000, 10000, 20000],
        metadata: {
          filename: file.name,
          filetype: file.type,
        },
        onError: (error) => {
          console.log(`Failed because: ${error}`);
        },
        onProgress: (bytesUploaded, bytesTotal) => {
          const percentage = (bytesUploaded / bytesTotal * 100).toFixed(2);
          console.log(bytesUploaded, bytesTotal, `${percentage}%`);
        },
        onSuccess: () => {
          console.log('Download %s from %s', upload.file.name, upload.url);
        },
      });

      // Start the upload
      upload.start();
    },
  },
};
</script>
