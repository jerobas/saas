import { EventEmitter } from 'events';

const sseEmitter = new EventEmitter();

export const sendSSE = (req, res) => {
  res.setHeader('Content-Type', 'text/event-stream');
  res.setHeader('Cache-Control', 'no-cache');
  res.setHeader('Connection', 'keep-alive');

  const clientId = req.params.clientId;

  const onUpdate = (data) => {
    if (data.clientId === clientId) {
      res.write(`data: ${JSON.stringify(data)}\n\n`);
    }
  };

  sseEmitter.on('update', onUpdate);

  req.on('close', () => {
    sseEmitter.removeListener('update', onUpdate);
  });
};

export const notifySSE = (data) => {
  sseEmitter.emit('update', data);
};