import { check } from 'k6';
import ws from 'k6/ws';
import { sleep } from 'k6';

export let options = {
  vus: 1,
  duration: '1m',
};

export default function () {
  const url = 'ws://localhost:8080/ws';
  const params = { tags: { my_tag: 'long_connection_test' } };

  const res = ws.connect(url, params, function (socket) {
    socket.on('open', function () {
      console.log('Connected');
      socket.send('Hello from client');
    });

    socket.on('message', function (message) {
      console.log('Received message: ' + message);
    });

    socket.on('close', function () {
      console.log('Connection closed');
    });

    socket.on('error', function (e) {
      console.log('WebSocket error: ' + e.error());
    });

    // 模拟客户端低频率发送消息，每5秒发送一次
    socket.setInterval(function () {
      socket.send('Low frequency message');
    }, 5000);

    // 保持连接300秒（5分钟）
    sleep(60);
    socket.close();
  });

  check(res, { 'status is 101': (r) => r && r.status === 101 });
}
