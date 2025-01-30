<script setup lang="ts">
import { ref } from 'vue'

const isBroadcasting = ref(false)
const isListening = ref(false)
const peerConnection = ref<RTCPeerConnection | null>(null)
const remoteAudioStream = ref<MediaStream | null>(null)
let localStream: MediaStream | null = null
let wsSignal: WebSocket | null = null

// WebRTC Configuration (STUN server for NAT traversal)
const rtcConfig = {
  iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
}

// Start broadcasting (WebRTC PeerConnection)
const startBroadcasting = async () => {
  localStream = await navigator.mediaDevices.getUserMedia({ audio: true })
  peerConnection.value = new RTCPeerConnection(rtcConfig)

  // Add microphone stream to PeerConnection
  localStream.getTracks().forEach(track => {
    peerConnection.value!.addTrack(track, localStream!)
  })

  // Set up WebSocket signaling to exchange WebRTC offers/answers
  wsSignal = new WebSocket('ws://localhost:8080/ws?mode=broadcast')
  
  wsSignal.onopen = () => {
    console.log('Signal server connected (broadcaster)')
    isBroadcasting.value = true
  }

  wsSignal.onmessage = async (message) => {
    const data = JSON.parse(message.data)

    if (data.type === 'answer') {
      console.log('Received answer:', data)
      await peerConnection.value!.setRemoteDescription(new RTCSessionDescription(data))
    }
  }

  // Generate WebRTC offer
  const offer = await peerConnection.value.createOffer()
  await peerConnection.value.setLocalDescription(offer)

  if (wsSignal.readyState === WebSocket.OPEN) wsSignal.send(JSON.stringify({ type: 'offer', offer }))
}

// Stop broadcasting
const stopBroadcasting = () => {
  peerConnection.value?.close()
  localStream?.getTracks().forEach(track => track.stop())
  wsSignal?.close()
  
  isBroadcasting.value = false
}

// Start listening (receiving WebRTC audio)
const startListening = () => {
  peerConnection.value = new RTCPeerConnection(rtcConfig)

  peerConnection.value.ontrack = (event) => {
    remoteAudioStream.value = event.streams[0]
    console.log('Received remote audio stream')
    
    // Play the received audio stream
    const audio = new Audio()
    audio.srcObject = remoteAudioStream.value
    audio.play()
  }

  // Connect to WebSocket signaling server
  wsSignal = new WebSocket('ws://localhost:8080/ws?mode=listen')

  wsSignal.onopen = () => {
    console.log('Signal server connected (listener)')
    isListening.value = true
  }

  wsSignal.onmessage = async (message) => {
    const data = JSON.parse(message.data)

    if (data.type === 'offer') {
      console.log('Received offer:', data)

      await peerConnection.value!.setRemoteDescription(new RTCSessionDescription(data.offer))
      const answer = await peerConnection.value!.createAnswer()
      await peerConnection.value!.setLocalDescription(answer)

      wsSignal?.send(JSON.stringify({ type: 'answer', answer }))
    }
  }
}

// Stop listening
const stopListening = () => {
  peerConnection.value?.close()
  remoteAudioStream.value = null
  wsSignal?.close()
  isListening.value = false
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
