import { OrbitControls } from '@react-three/drei'
import { Canvas, useFrame } from '@react-three/fiber'
import { useEffect, useMemo, useRef, useState } from 'react'
import { BufferAttribute, Color, DoubleSide, Group, Mesh, PlaneGeometry } from 'three'

export function LighthouseScene({ still }: { still: boolean }) {
  const reducedMotion = useReducedMotion()
  const frozen = still || reducedMotion

  return (
    <div className="lighthouse-canvas" data-scene-mode={still ? 'still' : 'live'}>
      <Canvas
        camera={{ position: [18, 13, 22], fov: 42, near: 0.1, far: 100 }}
        dpr={[1, 1.5]}
        shadows
        gl={{ antialias: true, powerPreference: 'high-performance' }}
        onCreated={({ gl }) => {
          gl.setClearColor(new Color('#071627'))
          gl.shadowMap.autoUpdate = !still
        }}
      >
        <fog attach="fog" args={['#071627', 20, 46]} />
        <hemisphereLight args={['#7180b5', '#07110f', 1.15]} />
        <ambientLight intensity={0.35} color="#7180b5" />
        <directionalLight
          castShadow
          position={[-8, 14, 8]}
          intensity={2.2}
          color="#d9e2ff"
          shadow-mapSize={[1024, 1024]}
          shadow-camera-far={38}
          shadow-camera-left={-14}
          shadow-camera-right={14}
          shadow-camera-top={14}
          shadow-camera-bottom={-14}
        />
        <pointLight position={[0, 8.8, 0]} intensity={12} distance={19} color="#ffe39a" />
        <Scene frozen={frozen} />
        <OrbitControls
          enabled={!still}
          enableDamping
          enablePan={false}
          minDistance={20}
          maxDistance={32}
          minPolarAngle={Math.PI * 0.27}
          maxPolarAngle={Math.PI * 0.43}
          minAzimuthAngle={-Math.PI * 0.32}
          maxAzimuthAngle={Math.PI * 0.32}
          target={[0, 3.1, 0]}
        />
      </Canvas>
    </div>
  )
}

function Scene({ frozen }: { frozen: boolean }) {
  return (
    <>
      <Stars />
      <Sea frozen={frozen} />
      <Island />
      <Lighthouse frozen={frozen} />
      <DecorativeFleet frozen={frozen} />
    </>
  )
}

function Stars() {
  const positions = useMemo(() => {
    const values: number[] = []
    let seed = 17
    for (let index = 0; index < 90; index += 1) {
      seed = (seed * 16807) % 2147483647
      const x = (seed / 2147483647 - 0.5) * 55
      seed = (seed * 16807) % 2147483647
      const y = 8 + (seed / 2147483647) * 19
      seed = (seed * 16807) % 2147483647
      const z = -12 - (seed / 2147483647) * 22
      values.push(x, y, z)
    }
    return new Float32Array(values)
  }, [])
  return (
    <points>
      <bufferGeometry>
        <bufferAttribute attach="attributes-position" args={[positions, 3]} />
      </bufferGeometry>
      <pointsMaterial color="#f5efcf" size={0.08} sizeAttenuation />
    </points>
  )
}

function Sea({ frozen }: { frozen: boolean }) {
  const mesh = useRef<Mesh<PlaneGeometry>>(null)
  const geometry = useMemo(() => {
    const value = new PlaneGeometry(48, 48, 48, 48)
    value.rotateX(-Math.PI / 2)
    return value
  }, [])
  const base = useMemo(
    () => Float32Array.from((geometry.attributes.position as BufferAttribute).array),
    [geometry],
  )

  useFrame(({ clock }) => {
    if (frozen || !mesh.current) return
    const position = mesh.current.geometry.attributes.position as BufferAttribute
    const elapsed = clock.elapsedTime
    for (let index = 0; index < position.count; index += 1) {
      const offset = index * 3
      const x = base[offset]
      const z = base[offset + 2]
      position.setY(
        index,
        Math.sin(x * 0.62 + elapsed * 0.55) * 0.12 +
          Math.cos(z * 0.48 - elapsed * 0.4) * 0.08,
      )
    }
    position.needsUpdate = true
    mesh.current.geometry.computeVertexNormals()
  })

  return (
    <mesh ref={mesh} geometry={geometry} receiveShadow position={[0, -0.35, 0]}>
      <meshStandardMaterial
        color="#0b3e4a"
        roughness={0.82}
        metalness={0.12}
        flatShading
        side={DoubleSide}
      />
    </mesh>
  )
}

function Island() {
  return (
    <group position={[0, -0.15, 0]}>
      <mesh receiveShadow castShadow scale={[1.2, 0.55, 1]}>
        <dodecahedronGeometry args={[4.2, 0]} />
        <meshStandardMaterial color="#39453e" roughness={1} flatShading />
      </mesh>
      <mesh receiveShadow position={[0, 1.25, 0]} scale={[1, 0.32, 0.9]}>
        <cylinderGeometry args={[3.1, 3.7, 1.2, 9]} />
        <meshStandardMaterial color="#736d50" roughness={1} flatShading />
      </mesh>
    </group>
  )
}

function Lighthouse({ frozen }: { frozen: boolean }) {
  const beam = useRef<Group>(null)
  useFrame(({ clock }) => {
    if (!frozen && beam.current)
      beam.current.rotation.y = clock.elapsedTime * (Math.PI / 6)
  })
  return (
    <group position={[0, 1.1, 0]}>
      <mesh castShadow receiveShadow position={[0, 2.65, 0]}>
        <cylinderGeometry args={[1.02, 1.45, 5.3, 12]} />
        <meshStandardMaterial color="#e7dfc5" roughness={0.9} flatShading />
      </mesh>
      <mesh castShadow position={[0, 5.45, 0]}>
        <cylinderGeometry args={[1.35, 1.35, 0.32, 12]} />
        <meshStandardMaterial color="#a33d2e" roughness={0.75} flatShading />
      </mesh>
      <mesh castShadow position={[0, 5.95, 0]}>
        <cylinderGeometry args={[0.9, 0.9, 0.85, 12]} />
        <meshStandardMaterial
          color="#ffe29a"
          emissive="#ba8a40"
          emissiveIntensity={1.5}
          transparent
          opacity={0.82}
        />
      </mesh>
      <mesh castShadow position={[0, 6.58, 0]}>
        <coneGeometry args={[1.25, 1.05, 12]} />
        <meshStandardMaterial color="#8f332b" roughness={0.78} flatShading />
      </mesh>
      <Windows />
      <group ref={beam} position={[0, 5.95, 0]} rotation={[0, -0.35, 0]}>
        <mesh position={[3.6, 0, 0]} rotation={[0, 0, -Math.PI / 2]}>
          <coneGeometry args={[0.82, 7.2, 24, 1, true]} />
          <meshBasicMaterial
            color="#ffe8a8"
            transparent
            opacity={0.11}
            side={DoubleSide}
            depthWrite={false}
          />
        </mesh>
      </group>
    </group>
  )
}

function Windows() {
  const levels = [2.05, 2.95, 3.85, 4.65, 5.28]
  return (
    <>
      {levels.map((height, index) => {
        const angle = index % 2 === 0 ? 0.3 : -0.42
        const radius = 1.06 - index * 0.035
        return (
          <mesh
            key={height}
            position={[Math.sin(angle) * radius, height, Math.cos(angle) * radius]}
            rotation={[0, angle, 0]}
          >
            <boxGeometry args={[0.38, 0.52, 0.08]} />
            <meshStandardMaterial
              color="#b9a968"
              emissive="#5d542e"
              emissiveIntensity={0.55}
            />
          </mesh>
        )
      })}
    </>
  )
}

function DecorativeFleet({ frozen }: { frozen: boolean }) {
  const boats = [
    { position: [-7, 0.05, 3] as [number, number, number], scale: 0.8, speed: 0.16 },
    { position: [7.5, -0.02, 1.2] as [number, number, number], scale: 0.64, speed: 0.12 },
    {
      position: [-5.7, -0.04, -5.2] as [number, number, number],
      scale: 0.55,
      speed: 0.1,
    },
  ]
  return (
    <>
      {boats.map((boat, index) => (
        <Boat key={index} {...boat} frozen={frozen} phase={index * 1.7} />
      ))}
    </>
  )
}

function Boat({
  position,
  scale,
  speed,
  phase,
  frozen,
}: {
  position: [number, number, number]
  scale: number
  speed: number
  phase: number
  frozen: boolean
}) {
  const group = useRef<Group>(null)
  useFrame(({ clock }) => {
    if (!frozen && group.current) {
      group.current.position.y =
        position[1] + Math.sin(clock.elapsedTime * 0.7 + phase) * 0.08
      group.current.rotation.z = Math.sin(clock.elapsedTime * 0.5 + phase) * 0.035
      group.current.position.x =
        position[0] + Math.sin(clock.elapsedTime * speed + phase) * 0.45
    }
  })
  return (
    <group ref={group} position={position} scale={scale} rotation={[0, phase - 0.4, 0]}>
      <mesh castShadow>
        <boxGeometry args={[2.3, 0.45, 0.8]} />
        <meshStandardMaterial color="#e7e4d5" roughness={0.8} />
      </mesh>
      <mesh castShadow position={[0.15, 0.47, 0]}>
        <boxGeometry args={[0.75, 0.5, 0.58]} />
        <meshStandardMaterial color="#f5f0df" roughness={0.85} />
      </mesh>
      <mesh position={[-0.15, 1.02, 0]}>
        <cylinderGeometry args={[0.035, 0.035, 1, 6]} />
        <meshStandardMaterial color="#ba8a40" />
      </mesh>
    </group>
  )
}

function useReducedMotion() {
  const [reduced, setReduced] = useState(false)
  useEffect(() => {
    const query = window.matchMedia('(prefers-reduced-motion: reduce)')
    const update = () => setReduced(query.matches)
    update()
    query.addEventListener('change', update)
    return () => query.removeEventListener('change', update)
  }, [])
  return reduced
}
