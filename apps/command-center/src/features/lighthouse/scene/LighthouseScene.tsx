import { OrbitControls } from '@react-three/drei'
import { Canvas, type ThreeEvent, useFrame } from '@react-three/fiber'
import { useEffect, useMemo, useRef, useState } from 'react'
import {
  BufferAttribute,
  BufferGeometry,
  Color,
  DoubleSide,
  Float32BufferAttribute,
  Group,
  Fog,
  HemisphereLight,
  AmbientLight,
  DirectionalLight,
  Mesh,
  MeshStandardMaterial,
  PlaneGeometry,
  PointsMaterial,
  Shape,
} from 'three'
import { findDestination } from '../destinations'
import {
  evaluatePerformanceWindow,
  FrameWindowMonitor,
  initialPerformanceState,
  PERFORMANCE_LIMITS,
  QUALITY_PROFILES,
  simulatedFps,
  type PerformanceSimulation,
  type SceneQuality,
} from '../performance'
import type {
  ApplicationVisualState,
  LighthouseVisualState,
  WeatherLevel,
} from '../lighthouseState'

interface LighthouseSceneProps {
  still: boolean
  visualState: LighthouseVisualState
  activeDestination: string | null
  onDestinationChange: (id: string | null) => void
  onNavigate: (route: string) => void
  quality: SceneQuality
  simulation: PerformanceSimulation
  onQualityChange: (quality: SceneQuality) => void
  onPerformance: (fps: number, remaining: number) => void
  onPerformanceFallback: () => void
  onFailure: () => void
  onReady: () => void
}

export function LighthouseScene({
  still,
  visualState,
  activeDestination,
  onDestinationChange,
  onNavigate,
  quality,
  simulation,
  onQualityChange,
  onPerformance,
  onPerformanceFallback,
  onFailure,
  onReady,
}: LighthouseSceneProps) {
  const reducedMotion = useReducedMotion()
  const frozen = still || reducedMotion
  const profile = QUALITY_PROFILES[quality]
  const [canvasKey, setCanvasKey] = useState(0)
  const restoredOnce = useRef(false)

  return (
    <div
      className={`lighthouse-canvas${activeDestination ? ' destination-active' : ''}`}
      data-scene-mode={still ? 'still' : 'live'}
      data-scene-quality={quality}
    >
      <Canvas
        key={canvasKey}
        camera={{ position: [18, 13, 22], fov: 42, near: 0.1, far: 100 }}
        dpr={quality === 'normal' ? [1, profile.dpr] : profile.dpr}
        shadows={profile.shadows}
        gl={{ antialias: true, powerPreference: 'high-performance' }}
        onPointerMissed={() => onDestinationChange(null)}
        onCreated={({ gl }) => {
          gl.setClearColor(new Color('#071627'))
          gl.shadowMap.autoUpdate = !still
          const canvas = gl.domElement
          let timer: ReturnType<typeof setTimeout> | undefined
          const lost = (event: Event) => {
            event.preventDefault()
            timer = setTimeout(onFailure, 3_000)
          }
          const restored = () => {
            if (timer) clearTimeout(timer)
            if (!restoredOnce.current) {
              restoredOnce.current = true
              setCanvasKey((value) => value + 1)
            } else {
              onFailure()
            }
          }
          canvas.addEventListener('webglcontextlost', lost, { once: true })
          canvas.addEventListener('webglcontextrestored', restored, { once: true })
          onReady()
        }}
      >
        <PerformanceMonitor
          simulation={simulation}
          onQualityChange={onQualityChange}
          onPerformance={onPerformance}
          onFallback={onPerformanceFallback}
        />
        <Atmosphere weather={visualState.weather} frozen={frozen} quality={quality} />
        <pointLight position={[0, 8.8, 0]} intensity={12} distance={19} color="#ffe39a" />
        <Scene
          frozen={frozen}
          visualState={visualState}
          activeDestination={activeDestination}
          onDestinationChange={onDestinationChange}
          onNavigate={onNavigate}
          quality={quality}
        />
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

interface InteractionProps {
  activeDestination: string | null
  onDestinationChange: (id: string | null) => void
  onNavigate: (route: string) => void
}

function Scene({
  frozen,
  visualState,
  quality,
  ...interaction
}: {
  frozen: boolean
  visualState: LighthouseVisualState
  quality: SceneQuality
} & InteractionProps) {
  return (
    <>
      <Stars
        weather={visualState.weather}
        frozen={frozen}
        count={QUALITY_PROFILES[quality].stars}
      />
      <CloudBanks
        weather={visualState.weather}
        frozen={frozen}
        count={QUALITY_PROFILES[quality].clouds}
      />
      <Sea
        frozen={frozen}
        weather={visualState.weather}
        segments={QUALITY_PROFILES[quality].seaSegments}
      />
      <Island />
      <Lighthouse
        frozen={frozen}
        applications={visualState.applications}
        {...interaction}
      />
      <NarrativeFleet frozen={frozen} {...interaction} />
      <DecorativeFleet frozen={frozen} count={QUALITY_PROFILES[quality].boats} />
    </>
  )
}

const weatherSettings = {
  clear: {
    sky: '#071627',
    fogNear: 22,
    fogFar: 48,
    stars: 0.94,
    cloud: 0.08,
    wave: 1,
    sea: '#ffffff',
    light: 2.2,
  },
  mist: {
    sky: '#152335',
    fogNear: 14,
    fogFar: 38,
    stars: 0.42,
    cloud: 0.24,
    wave: 1.18,
    sea: '#c3ccd0',
    light: 1.75,
  },
  storm: {
    sky: '#091827',
    fogNear: 13,
    fogFar: 36,
    stars: 0.14,
    cloud: 0.42,
    wave: 1.55,
    sea: '#87949d',
    light: 1.75,
  },
} as const

function Atmosphere({
  weather,
  frozen,
  quality,
}: {
  weather: WeatherLevel
  frozen: boolean
  quality: SceneQuality
}) {
  const fog = useRef<Fog>(null)
  const hemisphere = useRef<HemisphereLight>(null)
  const ambient = useRef<AmbientLight>(null)
  const directional = useRef<DirectionalLight>(null)
  const target = weatherSettings[weather]
  useFrame((state, delta) => {
    const factor = frozen ? 1 : 1 - Math.exp(-delta * 1.6)
    if (fog.current) {
      fog.current.color.lerp(new Color(target.sky), factor)
      fog.current.near += (target.fogNear - fog.current.near) * factor
      fog.current.far += (target.fogFar - fog.current.far) * factor
    }
    const background = state.gl.getClearColor(new Color())
    state.gl.setClearColor(background.lerp(new Color(target.sky), factor))
    if (hemisphere.current)
      hemisphere.current.intensity += (1.15 - hemisphere.current.intensity) * factor
    if (ambient.current)
      ambient.current.intensity +=
        ((weather === 'storm' ? 0.3 : 0.35) - ambient.current.intensity) * factor
    if (directional.current)
      directional.current.intensity +=
        (target.light - directional.current.intensity) * factor
  })
  return (
    <>
      <fog ref={fog} attach="fog" args={[target.sky, target.fogNear, target.fogFar]} />
      <hemisphereLight ref={hemisphere} args={['#7180b5', '#07110f', 1.15]} />
      <ambientLight ref={ambient} intensity={0.35} color="#7180b5" />
      <directionalLight
        ref={directional}
        castShadow={quality === 'normal'}
        position={[-8, 14, 8]}
        intensity={target.light}
        color="#d9e2ff"
        shadow-mapSize={[1024, 1024]}
        shadow-camera-far={38}
        shadow-camera-left={-14}
        shadow-camera-right={14}
        shadow-camera-top={14}
        shadow-camera-bottom={-14}
      />
    </>
  )
}

function Stars({
  weather,
  frozen,
  count,
}: {
  weather: WeatherLevel
  frozen: boolean
  count: number
}) {
  const material = useRef<PointsMaterial>(null)
  const positions = useMemo(() => {
    const values: number[] = []
    let seed = 17
    for (let index = 0; index < count; index += 1) {
      seed = (seed * 16807) % 2147483647
      const x = (seed / 2147483647 - 0.5) * 55
      seed = (seed * 16807) % 2147483647
      const y = 8 + (seed / 2147483647) * 19
      seed = (seed * 16807) % 2147483647
      const z = -12 - (seed / 2147483647) * 22
      values.push(x, y, z)
    }
    return new Float32Array(values)
  }, [count])
  useFrame((_, delta) => {
    if (!material.current) return
    const target = weatherSettings[weather].stars
    material.current.opacity +=
      (target - material.current.opacity) * (frozen ? 1 : 1 - Math.exp(-delta * 1.8))
  })
  return (
    <points raycast={() => null}>
      <bufferGeometry>
        <bufferAttribute attach="attributes-position" args={[positions, 3]} />
      </bufferGeometry>
      <pointsMaterial
        ref={material}
        color="#f5efcf"
        size={0.08}
        sizeAttenuation
        transparent
        opacity={weatherSettings[weather].stars}
      />
    </points>
  )
}

function Sea({
  frozen,
  weather,
  segments,
}: {
  frozen: boolean
  weather: WeatherLevel
  segments: number
}) {
  const mesh = useRef<Mesh<PlaneGeometry>>(null)
  const material = useRef<MeshStandardMaterial>(null)
  const wave = useRef(weatherSettings[weather].wave)
  const geometry = useMemo(() => {
    const value = new PlaneGeometry(48, 48, segments, segments)
    value.rotateX(-Math.PI / 2)
    const position = value.attributes.position as BufferAttribute
    const colors: number[] = []
    const deep = new Color('#082e3d')
    const light = new Color('#17606a')
    for (let index = 0; index < position.count; index += 1) {
      const x = position.getX(index)
      const z = position.getZ(index)
      position.setY(index, Math.sin(x * 0.62) * 0.12 + Math.cos(z * 0.48) * 0.08)
      const mix = (Math.sin(x * 0.55) + Math.cos(z * 0.43) + 2) / 4
      const color = deep.clone().lerp(light, mix * 0.72)
      colors.push(color.r, color.g, color.b)
    }
    value.setAttribute('color', new Float32BufferAttribute(colors, 3))
    value.computeVertexNormals()
    return value
  }, [segments])
  const base = useMemo(
    () => Float32Array.from((geometry.attributes.position as BufferAttribute).array),
    [geometry],
  )

  useFrame(({ clock }) => {
    if (!mesh.current) return
    const target = weatherSettings[weather]
    wave.current += (target.wave - wave.current) * (frozen ? 1 : 0.025)
    material.current?.color.lerp(new Color(target.sea), frozen ? 1 : 0.025)
    if (frozen) return
    const position = mesh.current.geometry.attributes.position as BufferAttribute
    const elapsed = clock.elapsedTime
    for (let index = 0; index < position.count; index += 1) {
      const offset = index * 3
      const x = base[offset]
      const z = base[offset + 2]
      position.setY(
        index,
        (Math.sin(x * 0.62 + elapsed * 0.55) * 0.12 +
          Math.cos(z * 0.48 - elapsed * 0.4) * 0.08) *
          wave.current,
      )
    }
    position.needsUpdate = true
    mesh.current.geometry.computeVertexNormals()
  })

  return (
    <>
      <mesh
        ref={mesh}
        geometry={geometry}
        receiveShadow
        position={[0, -0.35, 0]}
        raycast={() => null}
      >
        <meshStandardMaterial
          ref={material}
          color="#ffffff"
          vertexColors
          roughness={0.68}
          metalness={0.2}
          flatShading
          side={DoubleSide}
        />
      </mesh>
    </>
  )
}

function CloudBanks({
  weather,
  frozen,
  count,
}: {
  weather: WeatherLevel
  frozen: boolean
  count: number
}) {
  const material = useRef<MeshStandardMaterial>(null)
  useFrame((_, delta) => {
    if (material.current)
      material.current.opacity +=
        (weatherSettings[weather].cloud - material.current.opacity) *
        (frozen ? 1 : 1 - Math.exp(-delta * 1.4))
  })
  return (
    <group position={[0, 10, -17]} raycast={() => null}>
      {[-13, -7, 0, 8, 14].slice(0, count).map((x, index) => (
        <mesh key={x} position={[x, index % 2, -index]} scale={[5.2, 0.75, 1.6]}>
          <sphereGeometry args={[1, 10, 6]} />
          <meshStandardMaterial
            ref={index === 0 ? material : undefined}
            color="#8090a0"
            transparent
            opacity={weatherSettings[weather].cloud}
            depthWrite={false}
            roughness={1}
          />
        </mesh>
      ))}
    </group>
  )
}

function Island() {
  return (
    <group position={[0, -0.15, 0]}>
      <mesh receiveShadow castShadow scale={[1.2, 0.55, 1]} raycast={() => null}>
        <dodecahedronGeometry args={[4.2, 0]} />
        <meshStandardMaterial color="#39453e" roughness={1} flatShading />
      </mesh>
      <mesh
        receiveShadow
        position={[0, 1.25, 0]}
        scale={[1, 0.32, 0.9]}
        raycast={() => null}
      >
        <cylinderGeometry args={[3.1, 3.7, 1.2, 9]} />
        <meshStandardMaterial color="#736d50" roughness={1} flatShading />
      </mesh>
      <CoastalFoam />
      <CoastalRocks />
    </group>
  )
}

function CoastalFoam() {
  const patches = [
    [-4.3, -0.18, 0.2, 1.5],
    [-3.2, -0.18, 2.5, 1.1],
    [-1.1, -0.18, 3.5, 1.35],
    [1.9, -0.18, 3.2, 1.1],
    [4.1, -0.18, 1.1, 1.35],
    [4.2, -0.18, -1.3, 1.05],
    [2.2, -0.18, -3.1, 1.25],
    [-1.1, -0.18, -3.5, 1.15],
    [-3.7, -0.18, -2.1, 1.2],
  ] as const
  return (
    <group>
      {patches.map(([x, y, z, scale], index) => (
        <mesh
          key={index}
          position={[x, y, z]}
          scale={[scale, 0.035, 0.34]}
          rotation={[0, Math.atan2(z, x), 0]}
          raycast={() => null}
        >
          <sphereGeometry args={[0.65, 8, 5]} />
          <meshBasicMaterial
            color="#c9ddd5"
            transparent
            opacity={0.5}
            depthWrite={false}
          />
        </mesh>
      ))}
    </group>
  )
}

function CoastalRocks() {
  const rocks = [
    { position: [-5.2, 0.02, -1.4] as [number, number, number], scale: 0.62 },
    { position: [4.9, -0.05, 2.9] as [number, number, number], scale: 0.48 },
    { position: [2.7, -0.08, -4.6] as [number, number, number], scale: 0.72 },
    { position: [-3.8, -0.1, 4.1] as [number, number, number], scale: 0.42 },
  ]
  return (
    <group>
      {rocks.map((rock, index) => (
        <mesh
          key={index}
          position={rock.position}
          scale={[rock.scale, rock.scale * 0.48, rock.scale]}
          rotation={[0, index * 0.9, 0]}
          castShadow
          raycast={() => null}
        >
          <dodecahedronGeometry args={[1, 0]} />
          <meshStandardMaterial color="#303b39" roughness={1} flatShading />
        </mesh>
      ))}
    </group>
  )
}

function Lighthouse({
  frozen,
  applications,
  ...interaction
}: {
  frozen: boolean
  applications: LighthouseVisualState['applications']
} & InteractionProps) {
  const beam = useRef<Group>(null)
  useFrame(({ clock }) => {
    if (!frozen && beam.current)
      beam.current.rotation.y = clock.elapsedTime * (Math.PI / 6)
  })
  return (
    <group position={[0, 0.92, 0]} scale={1.13}>
      <mesh castShadow receiveShadow position={[0, 2.65, 0]} raycast={() => null}>
        <cylinderGeometry args={[1.02, 1.45, 5.3, 12]} />
        <meshStandardMaterial color="#e7dfc5" roughness={0.9} flatShading />
      </mesh>
      <mesh castShadow position={[0, 5.45, 0]} raycast={() => null}>
        <cylinderGeometry args={[1.35, 1.35, 0.32, 12]} />
        <meshStandardMaterial color="#a33d2e" roughness={0.75} flatShading />
      </mesh>
      <mesh castShadow position={[0, 5.95, 0]} raycast={() => null}>
        <cylinderGeometry args={[0.9, 0.9, 0.85, 12]} />
        <meshStandardMaterial
          color="#ffe29a"
          emissive="#ba8a40"
          emissiveIntensity={1.5}
          transparent
          opacity={0.82}
        />
      </mesh>
      <mesh castShadow position={[0, 6.58, 0]} raycast={() => null}>
        <coneGeometry args={[1.25, 1.05, 12]} />
        <meshStandardMaterial color="#8f332b" roughness={0.78} flatShading />
      </mesh>
      <Windows applications={applications} {...interaction} />
      <group ref={beam} position={[0, 5.95, 0]} rotation={[0, -0.35, 0]}>
        <mesh position={[3.6, 0, 0]} rotation={[0, 0, -Math.PI / 2]} raycast={() => null}>
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

function Windows({
  applications,
  ...interaction
}: { applications: LighthouseVisualState['applications'] } & InteractionProps) {
  const windows = [
    { id: 'farmami', height: 1.72, angle: -0.22 },
    { id: 'wheels_house', height: 2.55, angle: -0.1 },
    { id: 'prensap', height: 3.38, angle: 0.02 },
    { id: 'notizap', height: 4.18, angle: 0.14 },
    { id: 'atalaya', height: 4.94, angle: 0.26 },
  ]
  return (
    <>
      {windows.map((window) => {
        const radius = 1.48 - window.height * 0.081
        const position: [number, number, number] = [
          Math.sin(window.angle) * radius,
          window.height,
          Math.cos(window.angle) * radius,
        ]
        return (
          <group key={window.id} position={position} rotation={[0, window.angle, 0]}>
            <ArchitecturalWindow
              state={applications[window.id as keyof typeof applications]}
              active={interaction.activeDestination === window.id}
            />
            <DestinationHitbox
              id={window.id}
              size={[1.08, 0.92, 0.62]}
              {...interaction}
            />
          </group>
        )
      })}
    </>
  )
}

function ArchitecturalWindow({
  active,
  state,
}: {
  active: boolean
  state: ApplicationVisualState
}) {
  const shapes = useMemo(() => {
    const outer = archedWindowShape(0.62, 0.72)
    const inner = archedWindowShape(0.42, 0.55)
    return { outer, inner }
  }, [])
  const colors =
    state.severity === 'green'
      ? { glass: '#6fd69e', glow: '#198553' }
      : state.severity === 'red'
        ? { glass: '#ff7964', glow: '#c1432e' }
        : { glass: '#ffe478', glow: '#ba8a40' }
  const dimmed = state.freshness === 'stale'
  return (
    <group>
      <mesh raycast={() => null} position={[0, -0.08, 0]}>
        <extrudeGeometry
          args={[
            shapes.outer,
            {
              depth: 0.09,
              bevelEnabled: true,
              bevelSize: 0.035,
              bevelThickness: 0.025,
              bevelSegments: 2,
            },
          ]}
        />
        <meshStandardMaterial color="#8a6b34" roughness={0.66} metalness={0.28} />
      </mesh>
      <mesh raycast={() => null} position={[0.1, 0, 0.105]} scale={0.78}>
        <extrudeGeometry args={[shapes.inner, { depth: 0.04, bevelEnabled: false }]} />
        <meshStandardMaterial
          color={colors.glass}
          emissive={colors.glow}
          emissiveIntensity={(active ? 2 : 1.25) * (dimmed ? 0.48 : 1)}
          roughness={0.38}
        />
      </mesh>
      {dimmed && (
        <mesh raycast={() => null} position={[0.31, 0, 0.2]}>
          <torusGeometry args={[0.5, 0.025, 4, 10]} />
          <meshBasicMaterial color="#e0a13d" wireframe transparent opacity={0.9} />
        </mesh>
      )}
      <mesh raycast={() => null} position={[0.31, -0.42, 0.11]}>
        <boxGeometry args={[0.78, 0.1, 0.22]} />
        <meshStandardMaterial color="#9a783d" roughness={0.72} />
      </mesh>
      <mesh raycast={() => null} position={[0.31, 0.43, 0.08]} rotation={[0.18, 0, 0]}>
        <boxGeometry args={[0.75, 0.08, 0.28]} />
        <meshStandardMaterial color="#7d6134" roughness={0.78} />
      </mesh>
      <mesh raycast={() => null} position={[0.31, -0.02, 0.17]}>
        <boxGeometry args={[0.045, 0.61, 0.045]} />
        <meshStandardMaterial color="#5f4a2c" metalness={0.35} />
      </mesh>
    </group>
  )
}

function archedWindowShape(width: number, height: number) {
  const shape = new Shape()
  const radius = width / 2
  shape.moveTo(0, -height / 2)
  shape.lineTo(width, -height / 2)
  shape.lineTo(width, height / 2 - radius)
  shape.absarc(radius, height / 2 - radius, radius, 0, Math.PI, false)
  shape.lineTo(0, -height / 2)
  return shape
}

function NarrativeFleet({
  frozen,
  ...interaction
}: { frozen: boolean } & InteractionProps) {
  return (
    <>
      <PatrolBoat frozen={frozen} {...interaction} />
      <CargoShip frozen={frozen} {...interaction} />
      <MailBoat frozen={frozen} {...interaction} />
      <SystemBuoy frozen={frozen} {...interaction} />
    </>
  )
}

function PatrolBoat({ frozen, ...interaction }: { frozen: boolean } & InteractionProps) {
  const group = useRef<Group>(null)
  const light = useRef<Group>(null)
  useFrame(({ clock }) => {
    if (frozen) return
    const time = clock.elapsedTime
    if (group.current) {
      group.current.position.x = -10.2 + Math.sin(time * 0.24) * 0.72
      group.current.position.z = 4.6 + Math.cos(time * 0.24) * 0.38
      group.current.position.y = Math.sin(time * 0.8) * 0.07
    }
    if (light.current) light.current.rotation.y = Math.sin(time * 0.8) * 0.7
  })
  return (
    <group ref={group} position={[-10.2, 0, 4.6]} rotation={[0, -0.18, 0]} scale={1.02}>
      <Hull length={3.8} width={1.22} height={0.72} color="#d8dfe0" stripe="#2d5864" />
      <mesh castShadow position={[-0.2, 0.68, 0]} raycast={() => null}>
        <boxGeometry args={[1.35, 0.72, 0.94]} />
        <meshStandardMaterial color="#d9dedb" roughness={0.68} />
      </mesh>
      <mesh position={[-0.25, 0.78, 0.48]} raycast={() => null}>
        <boxGeometry args={[0.88, 0.28, 0.035]} />
        <meshStandardMaterial color="#18323d" roughness={0.25} />
      </mesh>
      <mesh position={[-0.25, 0.78, -0.48]} raycast={() => null}>
        <boxGeometry args={[0.88, 0.28, 0.035]} />
        <meshStandardMaterial color="#18323d" roughness={0.25} />
      </mesh>
      <mesh position={[0.48, 0.78, 0]} raycast={() => null}>
        <boxGeometry args={[0.035, 0.28, 0.72]} />
        <meshStandardMaterial color="#18323d" roughness={0.25} />
      </mesh>
      <RadarMast position={[-0.25, 1.48, 0]} />
      <Railings length={2.9} width={1.02} y={0.45} />
      <Wake position={[2.25, -0.3, 0]} scale={0.8} />
      <mesh castShadow position={[-1.52, 0.52, 0]} raycast={() => null}>
        <cylinderGeometry args={[0.18, 0.18, 0.38, 10]} />
        <meshStandardMaterial color="#c1432e" />
      </mesh>
      <group ref={light} position={[0.15, 1.26, 0]} rotation={[0, -0.4, 0]}>
        <mesh raycast={() => null} rotation={[0, 0, Math.PI / 2]}>
          <cylinderGeometry args={[0.15, 0.22, 0.3, 10]} />
          <meshStandardMaterial color="#d5c690" metalness={0.35} />
        </mesh>
        <mesh position={[1.4, 0, 0]} rotation={[0, 0, -Math.PI / 2]} raycast={() => null}>
          <coneGeometry args={[0.35, 2.8, 12, 1, true]} />
          <meshBasicMaterial
            color="#fff0b4"
            transparent
            opacity={0.22}
            side={DoubleSide}
            depthWrite={false}
          />
        </mesh>
      </group>
      <DestinationHitbox
        id="events"
        size={[4.8, 3.1, 2.5]}
        position={[0, 0.8, 0]}
        {...interaction}
      />
    </group>
  )
}

function CargoShip({ frozen, ...interaction }: { frozen: boolean } & InteractionProps) {
  const group = useRef<Group>(null)
  useFrame(({ clock }) => {
    if (!frozen && group.current) {
      const time = clock.elapsedTime
      group.current.position.x = 10.3 - ((time * 0.13) % 1.5)
      group.current.position.y = Math.sin(time * 0.55 + 1.2) * 0.055
    }
  })
  return (
    <group ref={group} position={[10.3, 0, 4.8]} rotation={[0, -0.65, 0]} scale={0.92}>
      <Hull length={5.3} width={1.5} height={0.82} color="#36454c" stripe="#8e3f30" />
      <mesh position={[-0.4, 0.48, 0]} raycast={() => null}>
        <boxGeometry args={[3.35, 0.16, 1.18]} />
        <meshStandardMaterial color="#6c6b5b" roughness={0.88} />
      </mesh>
      {[-1.25, -0.4, 0.45].map((x, index) => (
        <ContainerStack
          key={x}
          position={[x, 0.88, 0]}
          color={index === 1 ? '#a45338' : index === 2 ? '#566f63' : '#9a7a3f'}
        />
      ))}
      <mesh castShadow position={[1.72, 0.9, 0]} raycast={() => null}>
        <boxGeometry args={[1, 1.15, 1.08]} />
        <meshStandardMaterial color="#ddd8c7" roughness={0.8} />
      </mesh>
      <mesh position={[1.18, 1.15, 0.55]} raycast={() => null}>
        <boxGeometry args={[0.72, 0.3, 0.035]} />
        <meshStandardMaterial color="#203741" roughness={0.3} />
      </mesh>
      <mesh position={[1.18, 1.15, -0.55]} raycast={() => null}>
        <boxGeometry args={[0.72, 0.3, 0.035]} />
        <meshStandardMaterial color="#203741" roughness={0.3} />
      </mesh>
      <RadarMast position={[1.68, 2.05, 0]} />
      <Railings length={4.25} width={1.3} y={0.54} />
      <Wake position={[3.05, -0.34, 0]} scale={1.15} />
      <DestinationHitbox
        id="operations"
        size={[6.2, 3.5, 2.7]}
        position={[0, 0.9, 0]}
        {...interaction}
      />
    </group>
  )
}

function MailBoat({ frozen, ...interaction }: { frozen: boolean } & InteractionProps) {
  const group = useRef<Group>(null)
  useFrame(({ clock }) => {
    if (!frozen && group.current) {
      const time = clock.elapsedTime
      group.current.position.z = -3.4 + Math.sin(time * 0.18 + 2) * 0.52
      group.current.position.y = Math.sin(time * 0.68 + 2.4) * 0.065
      group.current.rotation.z = Math.sin(time * 0.48) * 0.025
    }
  })
  return (
    <group ref={group} position={[-12.1, 0, -3.4]} rotation={[0, 0.32, 0]} scale={0.9}>
      <Hull length={4.6} width={1.38} height={0.75} color="#e3ddc9" stripe="#9e3c31" />
      <mesh castShadow position={[0.28, 0.67, 0]} raycast={() => null}>
        <boxGeometry args={[2.25, 0.66, 1.12]} />
        <meshStandardMaterial color="#eee7d3" roughness={0.8} />
      </mesh>
      <mesh castShadow position={[0.48, 1.15, 0]} raycast={() => null}>
        <boxGeometry args={[1.62, 0.42, 0.96]} />
        <meshStandardMaterial color="#f2ead4" roughness={0.78} />
      </mesh>
      {[-0.15, 0.35, 0.85].map((x) => (
        <Porthole key={x} position={[x, 1.18, 0.495]} />
      ))}
      {[-0.15, 0.35, 0.85].map((x) => (
        <Porthole key={-x - 3} position={[x, 1.18, -0.495]} />
      ))}
      <mesh castShadow position={[-0.82, 1.28, 0]} raycast={() => null}>
        <cylinderGeometry args={[0.22, 0.26, 1.05, 12]} />
        <meshStandardMaterial color="#a23e31" roughness={0.76} />
      </mesh>
      <mesh position={[-0.82, 1.72, 0]} raycast={() => null}>
        <cylinderGeometry args={[0.24, 0.24, 0.16, 12]} />
        <meshStandardMaterial color="#20282b" />
      </mesh>
      <RadarMast position={[0.55, 1.95, 0]} />
      <Railings length={3.7} width={1.16} y={0.48} />
      <Wake position={[2.7, -0.32, 0]} scale={0.95} />
      <mesh position={[-1.5, 0.72, 0.7]} raycast={() => null}>
        <capsuleGeometry args={[0.13, 0.65, 4, 8]} />
        <meshStandardMaterial color="#d9c36a" />
      </mesh>
      <mesh position={[-1.5, 0.72, -0.7]} raycast={() => null}>
        <capsuleGeometry args={[0.13, 0.65, 4, 8]} />
        <meshStandardMaterial color="#d9c36a" />
      </mesh>
      <DestinationHitbox
        id="reports"
        size={[5.5, 3.6, 2.6]}
        position={[0, 0.95, 0]}
        {...interaction}
      />
    </group>
  )
}

function SystemBuoy({ frozen, ...interaction }: { frozen: boolean } & InteractionProps) {
  const group = useRef<Group>(null)
  const signal = useRef<Mesh>(null)
  useFrame(({ clock }) => {
    if (frozen) return
    const time = clock.elapsedTime
    if (group.current) {
      group.current.position.y = Math.sin(time * 0.72 + 0.8) * 0.09
      group.current.rotation.z = Math.sin(time * 0.55) * 0.045
    }
    if (signal.current) signal.current.visible = time % 2.4 < 0.55
  })
  return (
    <group ref={group} position={[9.2, 0.12, -4.4]} scale={1.18}>
      <mesh castShadow raycast={() => null}>
        <cylinderGeometry args={[0.62, 0.82, 0.72, 12]} />
        <meshStandardMaterial color="#c1432e" roughness={0.68} flatShading />
      </mesh>
      <mesh position={[0, -0.2, 0]} raycast={() => null}>
        <torusGeometry args={[0.77, 0.16, 8, 16]} />
        <meshStandardMaterial color="#d6d0ba" roughness={0.8} />
      </mesh>
      <mesh position={[0, 0.28, 0]} raycast={() => null}>
        <cylinderGeometry args={[0.7, 0.7, 0.16, 12]} />
        <meshStandardMaterial color="#efe8ce" roughness={0.76} />
      </mesh>
      {[-0.38, 0.38].map((x) => (
        <mesh key={x} position={[x, 1.1, 0]} raycast={() => null}>
          <cylinderGeometry args={[0.045, 0.055, 1.55, 8]} />
          <meshStandardMaterial color="#9b7a40" metalness={0.42} roughness={0.5} />
        </mesh>
      ))}
      <mesh position={[0, 1.82, 0]} raycast={() => null}>
        <torusGeometry args={[0.43, 0.055, 8, 16]} />
        <meshStandardMaterial color="#9b7a40" metalness={0.42} />
      </mesh>
      <mesh position={[0, 1.12, 0]} rotation={[0, 0, Math.PI / 2]} raycast={() => null}>
        <cylinderGeometry args={[0.04, 0.04, 0.82, 8]} />
        <meshStandardMaterial color="#9b7a40" metalness={0.42} />
      </mesh>
      {[0.65, 0.92, 1.19, 1.46].map((y) => (
        <mesh
          key={y}
          position={[-0.58, y, 0]}
          rotation={[0, 0, Math.PI / 2]}
          raycast={() => null}
        >
          <cylinderGeometry args={[0.025, 0.025, 0.32, 6]} />
          <meshStandardMaterial color="#c5b27a" metalness={0.28} />
        </mesh>
      ))}
      <mesh position={[0, 1.62, 0]} raycast={() => null}>
        <cylinderGeometry args={[0.16, 0.2, 0.3, 10]} />
        <meshStandardMaterial color="#24323a" metalness={0.3} />
      </mesh>
      <mesh ref={signal} position={[0, 1.78, 0]} raycast={() => null}>
        <sphereGeometry args={[0.15, 12, 8]} />
        <meshBasicMaterial color="#ffe478" />
      </mesh>
      <mesh position={[0, 1.78, 0]} raycast={() => null}>
        <cylinderGeometry args={[0.25, 0.25, 0.34, 10, 1, true]} />
        <meshStandardMaterial color="#8f7748" wireframe />
      </mesh>
      <DestinationHitbox
        id="system"
        size={[2.5, 3.8, 2.5]}
        position={[0, 0.9, 0]}
        {...interaction}
      />
    </group>
  )
}

function Hull({
  length,
  width,
  height,
  color,
  stripe,
}: {
  length: number
  width: number
  height: number
  color: string
  stripe: string
}) {
  const geometry = useMemo(
    () => createHullGeometry(length, width, height),
    [height, length, width],
  )
  return (
    <group>
      <mesh geometry={geometry} castShadow raycast={() => null}>
        <meshStandardMaterial color={color} roughness={0.72} flatShading />
      </mesh>
      <mesh position={[0.15, 0.18, 0]} scale={[0.92, 0.24, 1.02]} raycast={() => null}>
        <boxGeometry args={[length, height, width]} />
        <meshStandardMaterial color={stripe} roughness={0.8} />
      </mesh>
      <mesh
        geometry={geometry}
        position={[0, 0.27, 0]}
        scale={[0.94, 0.45, 0.94]}
        raycast={() => null}
      >
        <meshStandardMaterial color={color} roughness={0.76} flatShading />
      </mesh>
    </group>
  )
}

function createHullGeometry(length: number, width: number, height: number) {
  const sections = [
    { x: -length / 2, width: 0.04 },
    { x: -length * 0.28, width: width },
    { x: length * 0.3, width: width },
    { x: length / 2, width: width * 0.72 },
  ]
  const vertices: number[] = []
  for (const section of sections) {
    vertices.push(section.x, height / 2, section.width / 2)
    vertices.push(section.x, height / 2, -section.width / 2)
    vertices.push(section.x, -height / 2, section.width * 0.28)
    vertices.push(section.x, -height / 2, -section.width * 0.28)
  }
  const indices: number[] = []
  for (let index = 0; index < sections.length - 1; index += 1) {
    const a = index * 4
    const b = (index + 1) * 4
    indices.push(a, b, a + 1, b, b + 1, a + 1)
    indices.push(a + 2, a + 3, b + 2, b + 2, a + 3, b + 3)
    indices.push(a, a + 2, b, b, a + 2, b + 2)
    indices.push(a + 1, b + 1, a + 3, b + 1, b + 3, a + 3)
  }
  indices.push(0, 1, 2, 1, 3, 2)
  const last = (sections.length - 1) * 4
  indices.push(last, last + 2, last + 1, last + 1, last + 2, last + 3)
  const geometry = new BufferGeometry()
  geometry.setAttribute('position', new Float32BufferAttribute(vertices, 3))
  geometry.setIndex(indices)
  geometry.computeVertexNormals()
  return geometry
}

function RadarMast({ position }: { position: [number, number, number] }) {
  return (
    <group position={position}>
      <mesh raycast={() => null}>
        <cylinderGeometry args={[0.035, 0.045, 1, 8]} />
        <meshStandardMaterial color="#aa9159" metalness={0.35} />
      </mesh>
      <mesh position={[0, 0.38, 0]} rotation={[0, 0, Math.PI / 2]} raycast={() => null}>
        <boxGeometry args={[0.08, 0.65, 0.08]} />
        <meshStandardMaterial color="#d8d6c7" metalness={0.22} />
      </mesh>
      <mesh position={[0, 0.56, 0]} rotation={[Math.PI / 2, 0, 0]} raycast={() => null}>
        <torusGeometry args={[0.16, 0.025, 6, 12]} />
        <meshStandardMaterial color="#ba8a40" />
      </mesh>
    </group>
  )
}

function Railings({ length, width, y }: { length: number; width: number; y: number }) {
  const posts = [-length / 2, -length / 4, 0, length / 4, length / 2]
  return (
    <group>
      {[-width / 2, width / 2].map((z) => (
        <group key={z}>
          {posts.map((x) => (
            <mesh key={x} position={[x, y + 0.18, z]} raycast={() => null}>
              <cylinderGeometry args={[0.014, 0.014, 0.38, 5]} />
              <meshStandardMaterial color="#c9c8ba" metalness={0.38} />
            </mesh>
          ))}
          <mesh
            position={[0, y + 0.34, z]}
            rotation={[0, 0, Math.PI / 2]}
            raycast={() => null}
          >
            <cylinderGeometry args={[0.015, 0.015, length, 5]} />
            <meshStandardMaterial color="#c9c8ba" metalness={0.38} />
          </mesh>
        </group>
      ))}
    </group>
  )
}

function ContainerStack({
  position,
  color,
}: {
  position: [number, number, number]
  color: string
}) {
  return (
    <group position={position}>
      <mesh castShadow raycast={() => null}>
        <boxGeometry args={[0.76, 0.6, 1.08]} />
        <meshStandardMaterial color={color} roughness={0.84} />
      </mesh>
      {[-0.22, 0, 0.22].map((z) => (
        <mesh key={z} position={[0.39, 0, z]} raycast={() => null}>
          <boxGeometry args={[0.025, 0.5, 0.035]} />
          <meshStandardMaterial color="#322f29" />
        </mesh>
      ))}
    </group>
  )
}

function Porthole({ position }: { position: [number, number, number] }) {
  return (
    <mesh position={position} rotation={[Math.PI / 2, 0, 0]} raycast={() => null}>
      <cylinderGeometry args={[0.08, 0.08, 0.035, 12]} />
      <meshStandardMaterial
        color="#284654"
        emissive="#172a32"
        emissiveIntensity={0.45}
        metalness={0.18}
      />
    </mesh>
  )
}

function Wake({
  position,
  scale,
}: {
  position: [number, number, number]
  scale: number
}) {
  return (
    <group position={position} scale={scale}>
      {[0, 0.65, 1.3].map((x, index) => (
        <group key={x} position={[x, 0, 0]}>
          <mesh
            position={[0, 0, 0.36 + index * 0.12]}
            rotation={[0, 0.2, 0]}
            raycast={() => null}
          >
            <boxGeometry args={[1.35 + index * 0.35, 0.025, 0.075]} />
            <meshBasicMaterial
              color="#c5dcd7"
              transparent
              opacity={0.18 - index * 0.035}
              depthWrite={false}
            />
          </mesh>
          <mesh
            position={[0, 0, -0.36 - index * 0.12]}
            rotation={[0, -0.2, 0]}
            raycast={() => null}
          >
            <boxGeometry args={[1.35 + index * 0.35, 0.025, 0.075]} />
            <meshBasicMaterial
              color="#c5dcd7"
              transparent
              opacity={0.18 - index * 0.035}
              depthWrite={false}
            />
          </mesh>
        </group>
      ))}
    </group>
  )
}

function DestinationHitbox({
  id,
  size,
  position = [0, 0, 0],
  onDestinationChange,
  onNavigate,
}: {
  id: string
  size: [number, number, number]
  position?: [number, number, number]
} & InteractionProps) {
  function activate(event: ThreeEvent<MouseEvent>) {
    event.stopPropagation()
    if (event.delta > 4) return
    const destination = findDestination(id)
    if (destination) onNavigate(destination.route)
  }
  return (
    <mesh
      position={position}
      onClick={activate}
      onPointerOver={(event) => {
        event.stopPropagation()
        onDestinationChange(id)
      }}
      onPointerOut={(event) => {
        event.stopPropagation()
        onDestinationChange(null)
      }}
    >
      <boxGeometry args={size} />
      <meshBasicMaterial transparent opacity={0} depthWrite={false} />
    </mesh>
  )
}

function DecorativeFleet({ frozen, count }: { frozen: boolean; count: number }) {
  const boats = [
    {
      position: [11.8, -0.04, -7.8] as [number, number, number],
      scale: 0.42,
      speed: 0.08,
    },
    {
      position: [-13.2, -0.03, -7.5] as [number, number, number],
      scale: 0.38,
      speed: 0.07,
    },
    {
      position: [14.4, -0.06, 3.6] as [number, number, number],
      scale: 0.3,
      speed: 0.055,
    },
    {
      position: [-14.8, -0.05, 7.2] as [number, number, number],
      scale: 0.32,
      speed: 0.06,
    },
    {
      position: [3.4, -0.08, -11.5] as [number, number, number],
      scale: 0.28,
      speed: 0.045,
    },
    {
      position: [-3.8, -0.07, 11.8] as [number, number, number],
      scale: 0.26,
      speed: 0.05,
    },
  ]
  return (
    <>
      {boats.slice(0, count).map((boat, index) => (
        <DecorativeBoat key={index} {...boat} frozen={frozen} phase={index * 1.7} />
      ))}
    </>
  )
}

function PerformanceMonitor({
  simulation,
  onQualityChange,
  onPerformance,
  onFallback,
}: {
  simulation: PerformanceSimulation
  onQualityChange: (quality: SceneQuality) => void
  onPerformance: (fps: number, remaining: number) => void
  onFallback: () => void
}) {
  const monitor = useRef(new FrameWindowMonitor())
  const state = useRef(initialPerformanceState())
  useFrame(() => {
    const measured = monitor.current.frame(
      performance.now(),
      document.visibilityState === 'visible',
    )
    if (measured === null) return
    const next = evaluatePerformanceWindow(
      state.current,
      simulatedFps(simulation, measured),
    )
    state.current = next
    onQualityChange(next.quality)
    const remaining =
      next.quality === 'normal'
        ? PERFORMANCE_LIMITS.lowWindows - next.lowWindows
        : PERFORMANCE_LIMITS.criticalWindows - next.criticalWindows
    onPerformance(next.fps, Math.max(0, remaining))
    if (next.action === 'classic') onFallback()
  })
  return null
}

function DecorativeBoat({
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
    <group
      ref={group}
      position={position}
      scale={scale}
      rotation={[0, phase - 0.4, 0]}
      raycast={() => null}
    >
      <mesh castShadow raycast={() => null}>
        <boxGeometry args={[2.3, 0.45, 0.8]} />
        <meshStandardMaterial color="#e7e4d5" roughness={0.8} />
      </mesh>
      <mesh castShadow position={[0.15, 0.47, 0]} raycast={() => null}>
        <boxGeometry args={[0.75, 0.5, 0.58]} />
        <meshStandardMaterial color="#f5f0df" roughness={0.85} />
      </mesh>
      <Wake position={[1.55, -0.24, 0]} scale={0.42} />
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
