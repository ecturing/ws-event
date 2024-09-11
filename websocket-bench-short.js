import { check } from 'k6';
import ws from 'k6/ws';
import { sleep } from 'k6';

export let options = {
  stages: [
    { duration: '1m', target: 1000 },  // 在1分钟内将并发客户端提升到1000
    { duration: '3m', target: 1000 },  // 保持1000个并发客户端3分钟
    { duration: '1m', target: 0 },     // 在1分钟内将并发客户端减少到0
  ],
  thresholds: {
    'ws_duration': ['p(95)<2000'],     // 95%的响应时间应低于2秒
  },
};

export default function () {
  const url = 'ws://localhost:8080/ws';
  const params = { tags: { my_tag: 'websocket_stress_test' } };

  const res = ws.connect(url, params, function (socket) {
    socket.on('open', function () {
      console.log('Connected');
      // 每个连接开启后，立即发送一次消息
      socket.send('Initial message');
    });

    socket.on('message', function (message) {
      console.log('Received message: ' + message);
    });

    socket.on('close', function () {
      console.log('Disconnected');
    });

    socket.on('error', function (e) {
      console.log('WebSocket error: ' + e.error());
    });

    // 模拟客户端不断发送消息，每隔50ms发送一条
    socket.setInterval(function () {
      socket.send('High traffic message');
    }, 50);  // 每秒发送约20条消息

    // 保持连接10秒钟
    sleep(10);
    socket.close();
  });

  check(res, { 'status is 101': (r) => r && r.status === 101 });
}
