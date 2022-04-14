import React, { useCallback, useEffect, useRef, useMemo } from 'react';
import './App.css';
import config from './config';

const palette = ['#444', '#eee', '#f22', '#ee0', '#0e0', '#46f', '#f2f']
const pixelScale = 5;

export default function App() {
  const canvas = useRef({
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
  });

  const [selectedColor, setSelectedColor] = React.useState(1);

  const wsMessageHandler = useCallback((ev: MessageEvent<any>) => {
    const [y, x, color] = JSON.parse(ev.data);
    canvas.current.data[y * canvas.current.width + x] = color;
    canvas.current.draw();
  }, [canvas]);

  const ws = useMemo(() => {
    const url = new URL('/ws', config.apiUrl.replace(/^http/, 'ws'));
    const ws = new WebSocket(url);
    ws.onmessage = (e) => wsMessageHandler(e);
    return ws
  }, [wsMessageHandler]);

  const canvasEl = useRef<any>(null);

  useEffect(() => {
    const url = new URL('/canvas', config.apiUrl);
    fetch(url.toString()).then(async res => {
      const data: { pixels: number[] } = await res.json();
      canvas.current.data = data.pixels;
      canvas.current.draw()
    });
  }, []);

  useEffect(() => {
    const handleMousedown = (e: MouseEvent) => {
      // ピクセルを特定
      const bbox = canvasEl.current.getBoundingClientRect();
      const rx = e.clientX - bbox.left;
      const ry = e.clientY - bbox.top;
      const x = (rx / pixelScale | 0);
      const y = (ry / pixelScale | 0);

      if (x >= 0 && x < canvas.current.width && y >= 0 && y < canvas.current.height) {
        const url = new URL(`/canvas/pixels/${y}/${x}?color=${selectedColor}`, config.apiUrl);
        fetch(url.toString(), {
          method: 'POST',
        });
      }
    };

    window.addEventListener('mousedown', handleMousedown);
    return () => {
      window.removeEventListener('mousedown', handleMousedown);
    }
  }, [canvas, selectedColor]);

  return (
    <div className="App">
      <canvas id="canvas" ref={canvasEl}
        width={canvas.current.width * pixelScale} height={canvas.current.height * pixelScale}>
      </canvas>
      <div>
        {
          palette.map((color, i) => (
            <div
              style={{
                display: 'inline-block',
                margin: 3,
                padding: 2,
                width: 34,
                height: 34,
                background: selectedColor === i ? '#333' : '#fff',
                borderRadius: '50%'
              }}
              key={i}
              onClick={() => setSelectedColor(i)}>
              <div
                style={{
                  display: 'inline-block',
                  background: color,
                  width: 30,
                  height: 30,
                  borderRadius: '50%'
                }}
              ></div>
            </div>
          ))
        }
      </div>
    </div>
  );
}
