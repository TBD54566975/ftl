import { useContext, useState, useEffect } from 'react'
import { Config, Module, Secret, Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { FTLNode, GraphPane } from '../graph/GraphPane'
import { useClient } from '../../hooks/use-client'
import { Ref } from '../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { VerbService } from '../../protos/xyz/block/ftl/v1/ftl_connect'
import { Page } from '../../layout'
import { CubeTransparentIcon } from '@heroicons/react/24/outline'
import { modulesContext } from '../../providers/modules-provider'
import { NavigateFunction, useNavigate } from 'react-router-dom'
import { ResizablePanels } from '../../components/ResizablePanels'

const rowHeight = 100

const Fish = ({color}) => (
  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 576 512"><path fill={color} d="M180.5 141.5C219.7 108.5 272.6 80 336 80s116.3 28.5 155.5 61.5c39.1 33 66.9 72.4 81 99.8c4.7 9.2 4.7 20.1 0 29.3c-14.1 27.4-41.9 66.8-81 99.8C452.3 403.5 399.4 432 336 432s-116.3-28.5-155.5-61.5c-16.2-13.7-30.5-28.5-42.7-43.1L48.1 379.6c-12.5 7.3-28.4 5.3-38.7-4.9S-3 348.7 4.2 336.1L50 256 4.2 175.9c-7.2-12.6-5-28.4 5.3-38.6s26.1-12.2 38.7-4.9l89.7 52.3c12.2-14.6 26.5-29.4 42.7-43.1zM448 256a32 32 0 1 0 -64 0 32 32 0 1 0 64 0z"/></svg>
)

const Cat = ({color}) => (
  <svg style={{transform: 'scale(-1,1)'}} xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512"><path fill={color} d="M290.6 192c-20.2 0-106.8 2-162.6 86V192c0-52.9-43.1-96-96-96-17.7 0-32 14.3-32 32s14.3 32 32 32c17.6 0 32 14.4 32 32v256c0 35.3 28.7 64 64 64h176c8.8 0 16-7.2 16-16v-16c0-17.7-14.3-32-32-32h-32l128-96v144c0 8.8 7.2 16 16 16h32c8.8 0 16-7.2 16-16V289.9c-10.3 2.7-20.9 4.5-32 4.5-61.8 0-113.5-44.1-125.4-102.4zM448 96h-64l-64-64v134.4c0 53 43 96 96 96s96-43 96-96V32l-64 64zm-72 80c-8.8 0-16-7.2-16-16s7.2-16 16-16 16 7.2 16 16-7.2 16-16 16zm80 0c-8.8 0-16-7.2-16-16s7.2-16 16-16 16 7.2 16 16-7.2 16-16 16z"/></svg>
)

const PooIcon = ({color}) => (
  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512"><path fill={color} d="M254.4 6.6c3.5-4.3 9-6.5 14.5-5.7C315.8 7.2 352 47.4 352 96c0 11.2-1.9 22-5.5 32H352c35.3 0 64 28.7 64 64c0 19.1-8.4 36.3-21.7 48H408c39.8 0 72 32.2 72 72c0 23.2-11 43.8-28 57c34.1 5.7 60 35.3 60 71c0 39.8-32.2 72-72 72H72c-39.8 0-72-32.2-72-72c0-35.7 25.9-65.3 60-71c-17-13.2-28-33.8-28-57c0-39.8 32.2-72 72-72h13.7C104.4 228.3 96 211.1 96 192c0-35.3 28.7-64 64-64h16.2c44.1-.1 79.8-35.9 79.8-80c0-9.2-1.5-17.9-4.3-26.1c-1.8-5.2-.8-11.1 2.8-15.4z"/></svg>
)

const LilFish = ({color, col}) => {
  const [shouldRender, setShouldRender] = useState(true)
  setTimeout(() => setShouldRender(false), 1000)
  if (!shouldRender) {
    return []
  }

  return (
    <div style={{
      position: 'absolute',
      marginTop: '30px',
      marginLeft: 60 + 100*(col),
      width: '30px',
      height: '30px',
      animationName: `animationFish${col}`,
      animationDuration: '0.8s',
      animationDelay: '0.0s',
      animationIterationCount: 1,
      animationDirection: "normal",
      animationFillMode: "forwards"
    }}>
      <Fish color={color} />
    </div>
  )
}

const Editor = ({req, setMs, close}) => {
  const [msVal, setMsVal] = useState(req.ms)
  const modalBg = {
    position: 'absolute',
    backgroundColor: 'rgba(0,0,0,0.4)',
    top: 0,
    left: 0,
    width: '100vw',
    height: '100vh',
  }
  const modal = {
    borderRadius: '8px',
    position: 'absolute',
    backgroundColor: 'rgba(255,255,255,1)',
    opacity: 1,
    width: 250,
    top: '25vh',
    left: 'calc(50vw)',
    padding: '10px 0 40px 0',
  }
  const onChange = (e) => {
    setMsVal(e.target.value)
    req.ms = e.target.value
  }
  const submit = (e) => {
    setMs(msVal)
    close(e)
  }
  const onKeyDown = (e) => {
    if (e.key === 'Enter') {
      submit(e)
    }
  }

  return (
    <div style={modalBg}>
      <div style={modal}>
        <span className='text-lg'
          style={{marginLeft: 30}}
        >Set Call Interval (ms)</span>
        <hr style={{margin: '10px 0 30px 0'}} />
        <input name='ms' type='number' autoFocus
          style={{width: 100, marginLeft: 30, borderRadius: 8}}
          value={msVal}
          onChange={onChange}
          onKeyDown={onKeyDown}
        />
        <button
          className='bg-indigo-700 text-white ml-2 px-4 py-2 rounded-lg hover:bg-indigo-600 focus:outline-none focus:bg-indigo-600'
          onClick={submit}
        >
          OK
        </button>
      </div>
    </div>
  )
}

function timeoutPromise(ms) {
    return new Promise((r) => setTimeout(r, ms))
}

const FishBlock = ({req, color, col, callVerb}) => {
  const [ms, setMs] = useState(req.ms)
  const [lilFish, setLilFish] = useState([])
  const [showEditor, setShowEditor] = useState(false)
  const addLilFish = async () => {
    const key = `${Date.now()}`
    setLilFish([...lilFish, <LilFish key={key} color={color} col={col} />])
    await timeoutPromise(700)
    callVerb(req.req)
  }

  useEffect(() => {
    const interval = setInterval(addLilFish, ms)
    return () => clearInterval(interval)
  }, [lilFish, req, color])

  const onClick = (e) => {
    e.stopPropagation()
    if (!e.shiftKey) {
      return addLilFish()
    }
    setShowEditor(true)
  }
  const close = (e) => {
    e.stopPropagation()
    setShowEditor(false)
  }

  return [
    (
      <div key='bigfish'
        style={{display: 'inline-block', float: 'left', width: 80, margin: '5px 10px'}}
        onClick={onClick}
      >
        <Fish color={color} />
      </div>
    ),
    ...lilFish,
    showEditor ? (<Editor key='editor' req={req} setMs={setMs} close={close} />) : null,
  ]
}

const Poo = ({color}) => {
  const [shouldRender, setShouldRender] = useState(true)
  setTimeout(() => setShouldRender(false), 3000)
  if (!shouldRender) {
    return []
  }

  return (
    <div style={{
      position: 'absolute',
      marginTop: '-20px',
      marginLeft: 60,
      width: '20px',
      height: '20px',
      animationName: `animationPoo${Math.floor(Math.random()*4)}`,
      animationDuration: '2s',
      animationDelay: '0.0s',
      animationIterationCount: 1,
      animationDirection: "normal",
      animationFillMode: "forwards"
    }}>
      <PooIcon color={color} />
    </div>
  )
}

const CatBlock = ({verbRef, poos}) => {
  return (
    <div style={{float: 'right', width: 80, margin: '10px 10px'}}>
      <Cat color='#fa0' />
      {poos}
    </div>
  );
}

const blues = ['#03045e', '#023e8a', '#0077b6', '#0096c7', '#00b4d8', '#1576bb']

const Row = ({verbRef, callVerb}) => {
  const verbRefParts = verbRef.split('.')
  const verbRefPb: Ref = {
    name: verbRefParts[1],
    module: verbRefParts[0],
  } as Ref
  const [poos, setPoos] = useState([])
  const [lastErr, setLastErr] = useState('')
  const addPoo = (color) => {
    const key = `${Date.now()}`
    setPoos([...poos, <Poo key={key} color={color} />])
  }
  const addGoodPoo = () => addPoo('green')
  const addBadPoo = () => addPoo('red')
  const callVerbFn = async (req) => {
    const maybeErr = await callVerb(verbRefPb, req)
    if (maybeErr === null) {
        addGoodPoo()
    } else {
        addBadPoo()
        setLastErr(maybeErr)
    }
  }
  const fishes = window.savedReqs[verbRef].map((req, i) => (
    <FishBlock
      key={i} col={i}
      req={req}
      color={blues[i % blues.length]}
      callVerb={callVerbFn}
    />
  ))
  const titleStyle = {
    position: 'absolute',
    padding: '2px 20px',
  }
  const errStyle = {
      position: 'absolute',
      color: 'red',
      marginTop: 40,
      marginLeft: 100 * fishes.length + 20
  }
  const maybeErrEl = lastErr == '' ? [] : (
      <div style={errStyle}>{lastErr}</div>
  )
  return (
    <div style={{width: '100%', height: rowHeight}}>
      <div className='text-lg' style={titleStyle}>{verbRef}</div>
      {maybeErrEl}
      <div
        style={{display: 'inline-block', float: 'left', width: 'calc(100% - 120px)', margin: '10px 10px'}}
      >
        {fishes}
      </div>
      <CatBlock verbRef={verbRef} poos={poos} />
    </div>
  )
}

export const LoadTestPage = () => {
  const savedReqs = window.savedReqs ? window.savedReqs : []

  const client = useClient(VerbService)
  const modules = useContext(modulesContext)
  const navigate = useNavigate()
  const [selectedNode, setSelectedNode] = useState<FTLNode | null>(null)

  const callVerb = async (verbRef, req) => {
    const response = await client.call({ verb: verbRef, body: req })
    if (response.response.case === 'body') {
      const jsonString = Buffer.from(response.response.value).toString('utf-8')
      return null
    } else if (response.response.case === 'error') {
      return response.response.value.message
    }
  }

  const rows = Object.keys(savedReqs).toSorted().map((ref) => (
    <Row key={ref} verbRef={ref} callVerb={callVerb} />
  ))

  const gridStyle = {
    display: 'block',
    width: '100%',
    height: rowHeight * Object.keys(savedReqs).length,
    backgroundImage: "repeating-linear-gradient(#ccc 0 1px, transparent 1px 100%)",
    backgroundSize: `${rowHeight}px ${rowHeight}px`,
  }

  return (
    <Page>
      <Page.Header icon={<CubeTransparentIcon />} title='Console' />
      <Page.Body className='flex h-full'>
        <div style={gridStyle}>
          {rows}
        </div>
      </Page.Body>
    </Page>
  )
}
