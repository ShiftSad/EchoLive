<script setup lang="ts">
import { ref } from 'vue'

const isBroadcasting = ref(false)
const isListening = ref(false)

const broadcastingSocket = ref<WebSocket | null>(null)
const listeningSocket = ref<WebSocket | null>(null)

let peerConnection: RTCPeerConnection;

const startBroadcasting = async () => {
  const pendingCandidates: RTCIceCandidate[] = []

  broadcastingSocket.value = new WebSocket('ws://localhost:8080/ws')
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
      broadcastingSocket.value?.send(JSON.stringify({ 
        type: 'candidate',
        candidate: event.candidate 
      }))
    }
  }

  const offer = await peerConnection.createOffer();
  await peerConnection.setLocalDescription(offer);
  broadcastingSocket.value?.send(JSON.stringify({ 
    type: 'offer',
    offer 
  }))

  if (broadcastingSocket.value) {
    broadcastingSocket.value.onmessage = async (event) => {
      const data = JSON.parse(event.data);
      
      console.log('Received data:', data);
      if (data.answer) {
        // Only set the remote answer if we actually have a local offer
        if (peerConnection.signalingState === 'have-local-offer') {
          console.log('Setting remote description')
          await peerConnection.setRemoteDescription(new RTCSessionDescription(data.answer))
          for (const candidate of pendingCandidates) {
            await peerConnection.addIceCandidate(candidate)
          }
        } else {
          console.warn('Ignoring remote answer, signalingState:', peerConnection.signalingState)
        }
      } else if (data.candidate) {
        const candidate = new RTCIceCandidate(data.candidate)
        if (peerConnection.remoteDescription) {
          await peerConnection.addIceCandidate(candidate)
        } else pendingCandidates.push(candidate)
        console.log('Received candidate:', candidate)
      }
    }
  };
}

const startListening = async () => {
  listeningSocket.value = new WebSocket('ws://localhost:8080/ws')
  isListening.value = true

  // Create a new RTCPeerConnection for listening
  peerConnection = new RTCPeerConnection({
    iceServers: [{ urls: "stun:stun.l.google.com:19302" }]
  })

  // Receive audio track
  peerConnection.ontrack = (event) => {
    const [remoteStream] = event.streams
    const audio = new Audio()
    audio.srcObject = remoteStream
    audio.play()
  }

  // Send ICE candidates to server
  peerConnection.onicecandidate = (event) => {
    if (event.candidate) {
      listeningSocket.value?.send(JSON.stringify({ 
        type: 'candidate',
        candidate: event.candidate 
      }))
    }
  }

  // Listen for offer and respond with an answer
  if (listeningSocket.value) {
    listeningSocket.value.onmessage = async (event) => {
      const data = JSON.parse(event.data)

      if (data.offer) {
        await peerConnection.setRemoteDescription(new RTCSessionDescription(data.offer))
        const answer = await peerConnection.createAnswer()
        await peerConnection.setLocalDescription(answer)
        listeningSocket.value?.send(JSON.stringify({ type: 'answer', answer }))
      } else if (data.candidate) {
        const candidate = new RTCIceCandidate(data.candidate)
        if (peerConnection.remoteDescription) {
          await peerConnection.addIceCandidate(candidate)
        }
      }
    }
  }
}

const stopBroadcasting = () => {
  if (broadcastingSocket.value && broadcastingSocket.value.readyState === WebSocket.OPEN) {
    broadcastingSocket.value.close()
    listeningSocket.value = null
  }
  isBroadcasting.value = false
  peerConnection?.close()
}

const stopListening = () => {
  if (listeningSocket.value && listeningSocket.value.readyState === WebSocket.OPEN) {
    listeningSocket.value.close()
    listeningSocket.value = null
  }
  isListening.value = false
  peerConnection?.close()
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
