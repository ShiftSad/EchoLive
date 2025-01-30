<script setup lang="ts">
import { ref } from 'vue'

const isBroadcasting = ref(false)
const isListening = ref(false)

const broadcastingSocket = ref<WebSocket | null>(null)
const listeningSocket = ref<WebSocket | null>(null)

let peerConnection: RTCPeerConnection;

const startBroadcasting = async () => {
  broadcastingSocket.value = new WebSocket('ws://localhost:8080/ws?role=broadcast')
  isBroadcasting.value = true
  
  const stream = await navigator.mediaDevices.getUserMedia({ audio: true })

  peerConnection = new RTCPeerConnection({
    iceServers: [
      { "urls": "stun:stun.l.google.com:19302" }
    ],
  })

  stream.getTracks().forEach(track => peerConnection.addTrack(track, stream))

  peerConnection.onicecandidate = (event) => {
    if (event.candidate) {
      broadcastingSocket.value?.send(JSON.stringify({ candidate: event.candidate }));
    }
  }

  const offer = await peerConnection.createOffer();
  await peerConnection.setLocalDescription(offer);
  broadcastingSocket.value?.send(JSON.stringify({ offer }));
}

</script>

<template>
  <div class="container text-center mt-5">
    <h1 class="mb-4 text-primary">🎤 EchoLive - WebRTC Karaoke</h1>

    <div class="d-flex justify-content-center gap-3">
      <button v-if="!isBroadcasting" @click="startBroadcasting" class="btn btn-success">
        🎙 Start Broadcasting
      </button>

      <button v-if="isBroadcasting" @click="stopBroadcasting" class="btn btn-danger">
        ⛔ Stop Broadcasting
      </button>
    </div>

    <div class="mt-4 d-flex justify-content-center gap-3">
      <button v-if="!isListening" @click="startListening" class="btn btn-primary">
        🎧 Listen (Opt-in)
      </button>

      <button v-if="isListening" @click="stopListening" class="btn btn-warning">
        🔇 Stop Listening
      </button>
    </div>
  </div>
</template>

<style scoped>
.container {
  max-width: 600px;
}
</style>
