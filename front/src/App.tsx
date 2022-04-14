import React, { useCallback, useEffect, useRef, useMemo } from 'react';
import './App.css';

const palette = ['#444', '#eee', '#f22', '#6f0', '#44f']
const pixelScale = 5;

export default function App() {
  const canvas = useMemo(() => ({
    width: 100,
    height: 100,
    data: [...Array(100 * 100)].map((_, i) => i % 5),
    draw() {
      if (canvasEl.current instanceof HTMLCanvasElement) {
        const ctx = canvasEl.current.getContext('2d');
        if (ctx) {
          // draw background
          ctx.fillStyle = '#eee';
          ctx.fillRect(0, 0, this.width * pixelScale, this.height * pixelScale);

          // draw pixels
          for (let y = 0; y < this.height; ++y) {
            for (let x = 0; x < this.width; ++x) {
              const p = this.data[y * this.width + x];
              ctx.fillStyle = palette[p];
              ctx.fillRect(x * pixelScale, y * pixelScale, pixelScale, pixelScale);
            }
          }
        }
      }
    }
  }), []);

  const [selectedColor, setSelectedColor] = React.useState(1);

  const wsMessageHandler = useCallback((ev: MessageEvent<any>) => {
    const [y, x, color] = JSON.parse(ev.data);
    canvas.data[y * canvas.width + x] = color;
    canvas.draw();
  },[canvas]);

  const ws = useMemo(() => {
    const ws = new WebSocket('ws://localhost:8080/ws');
    ws.onmessage = (e) => wsMessageHandler(e);
    return ws
  }, []);

  const canvasEl = useRef<any>(null);

  useEffect(() => {
    fetch('http://localhost:8080/canvas').then(async res => {
      const data: {pixels: number[]} = await res.json();
      canvas.data = data.pixels;
      canvas.draw()
    });
  }, []);

  useEffect(() => {
    const handleMousedown = (e: MouseEvent)=> {
      // ピクセルを特定
      const bbox = canvasEl.current.getBoundingClientRect();
      const rx = e.clientX - bbox.left;
      const ry = e.clientY - bbox.top;
      const x = (rx / pixelScale | 0);
      const y = (ry / pixelScale | 0);

      if (x >= 0 && x < canvas.width && y >= 0 && y < canvas.height) {
        fetch(`http://localhost:8080/canvas/pixels/${y}/${x}?color=${selectedColor}`, {
          method: 'POST',
        });
      }
    };

    window.addEventListener('mousedown', handleMousedown);
    return () => {
      window.removeEventListener('mousedown', handleMousedown);
    }
  }, [selectedColor])

  return (
    <div className="App">
      <canvas id="canvas" ref={canvasEl}
      width={canvas.width * pixelScale} height={canvas.height * pixelScale}></canvas>
      <div>
        {
          palette.map((color, i) => (
            <div
            style={{
              display: 'inline-block',
              background: color,
              width: 30,
              height: 30,
              margin: 5,
              border: selectedColor === i ? '2px solid #444' : '2px solid #fff',
            }}
            key={i}
            onClick={() => setSelectedColor(i)}>
            </div>
          ))
        }
      </div>
    </div>
  );
}
