import { useMemo, useState } from 'react';
import {
  Alert,
  Button,
  Card,
  Col,
  Divider,
  Row,
  Select,
  Space,
  Table,
  Tag,
  Typography,
  message,
} from 'antd';

import type { SetTemplate, XraySettingsValue } from '@/hooks/useXraySetting';

import {
  EXIT_POOL_TAG_PREFIX,
  bindExitPoolRoute,
  listExitPoolBindings,
  unbindExitPoolRoute,
} from './exit-pool-routing';

interface ExitPoolTabProps {
  templateSettings: XraySettingsValue | null;
  setTemplateSettings: SetTemplate;
  inboundTags: string[];
}

interface ExitPoolOutbound {
  tag: string;
  address: string;
  port: number | null;
}

function exitPoolOutbounds(template: XraySettingsValue | null): ExitPoolOutbound[] {
  const rows: ExitPoolOutbound[] = [];
  for (const outbound of template?.outbounds || []) {
    const outboundRecord = outbound as unknown as Record<string, unknown>;
    const tag = typeof outboundRecord.tag === 'string' ? outboundRecord.tag : '';
    if (!tag.startsWith(EXIT_POOL_TAG_PREFIX)) continue;
    const settings = outboundRecord.settings as Record<string, unknown> | undefined;
    const servers = Array.isArray(settings?.servers) ? settings.servers : [];
    const server = (servers[0] || {}) as Record<string, unknown>;
    rows.push({
      tag,
      address: typeof server.address === 'string' ? server.address : '127.0.0.1',
      port: typeof server.port === 'number' ? server.port : null,
    });
  }
  return rows;
}

export default function ExitPoolTab({
  templateSettings,
  setTemplateSettings,
  inboundTags,
}: ExitPoolTabProps) {
  const [messageApi, messageContextHolder] = message.useMessage();
  const [inboundTag, setInboundTag] = useState('');
  const [outboundTag, setOutboundTag] = useState('');
  const outbounds = useMemo(() => exitPoolOutbounds(templateSettings), [templateSettings]);
  const bindings = useMemo(
    () => (templateSettings ? listExitPoolBindings(templateSettings) : []),
    [templateSettings],
  );

  function bind() {
    if (!templateSettings || !inboundTag || !outboundTag) {
      messageApi.warning('请选择入站和出口');
      return;
    }
    setTemplateSettings(bindExitPoolRoute(templateSettings, inboundTag, outboundTag));
    messageApi.success('绑定已加入配置，请点击页面顶部“保存”使其生效');
  }

  function unbind(tag: string) {
    if (!templateSettings) return;
    setTemplateSettings(unbindExitPoolRoute(templateSettings, tag));
    messageApi.success('绑定已移除，请点击页面顶部“保存”使其生效');
  }

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      {messageContextHolder}
      <Alert
        showIcon
        type="info"
        title="多国出口池是可选功能"
        description="出口池未安装、未启动或未绑定路由时，不会改变现有节点。出口由独立服务建立 OpenVPN 隧道并检测真实 IP；3x-ui 只管理 fanout-* 出站和显式路由绑定。"
      />

      <Card size="small" title="一键安装与管理">
        <Typography.Paragraph>
          在 VPS 的 root 终端执行安装命令。安装后用 access
          命令打开出口池控制台，在那里选择国家、出口数量、检测并发、本机直出或第三方 SOCKS5。
        </Typography.Paragraph>
        <Typography.Paragraph>
          <Typography.Link
            href="https://github.com/jrangus/3x-ui/blob/main/docs/exit-pool.md"
            target="_blank"
            rel="noreferrer"
          >
            查看完整部署与安全说明
          </Typography.Link>
        </Typography.Paragraph>
        <Space direction="vertical" style={{ width: '100%' }}>
          <Typography.Text code copyable>
            x-ui exit-pool install
          </Typography.Text>
          <Typography.Text code copyable>
            x-ui exit-pool access
          </Typography.Text>
          <Typography.Text code copyable>
            x-ui exit-pool update
          </Typography.Text>
          <Typography.Text code copyable>
            x-ui exit-pool status
          </Typography.Text>
        </Space>
      </Card>

      <Card size="small" title="已发现的出口">
        {outbounds.length === 0 ? (
          <Alert
            showIcon
            type="warning"
            title="尚未发现 fanout-* 出站"
            description="先安装并打开出口池控制台，创建出口并等待状态变为可用，然后刷新本页面。"
          />
        ) : (
          <Table
            size="small"
            pagination={false}
            rowKey="tag"
            dataSource={outbounds}
            columns={[
              { title: '出站标签', dataIndex: 'tag' },
              { title: '本地 SOCKS 地址', render: (_, row) => `${row.address}:${row.port ?? '-'}` },
              { title: '状态', render: () => <Tag color="green">已注册</Tag> },
            ]}
          />
        )}
      </Card>

      <Card size="small" title="绑定入站到出口">
        <Alert
          showIcon
          type="warning"
          title="第一版仅绑定 TCP"
          description="VPN Gate 公共出口当前不能保证 UDP、QUIC、Hysteria2 或 IPv6。绑定规则明确限制为 TCP，避免 UDP 静默从 VPS 本机直出。"
          style={{ marginBottom: 16 }}
        />
        <Row gutter={[12, 12]}>
          <Col xs={24} md={9}>
            <Select
              showSearch
              value={inboundTag || undefined}
              placeholder="选择要使用出口池的入站"
              options={inboundTags.map((tag) => ({ label: tag, value: tag }))}
              onChange={setInboundTag}
              style={{ width: '100%' }}
            />
          </Col>
          <Col xs={24} md={9}>
            <Select
              showSearch
              value={outboundTag || undefined}
              placeholder="选择 fanout 出站"
              options={outbounds.map((outbound) => ({ label: outbound.tag, value: outbound.tag }))}
              onChange={setOutboundTag}
              style={{ width: '100%' }}
            />
          </Col>
          <Col xs={24} md={6}>
            <Button type="primary" block onClick={bind}>
              创建绑定规则
            </Button>
          </Col>
        </Row>

        <Divider />
        <Table
          size="small"
          pagination={false}
          rowKey={(row) => `${row.inboundTag}:${row.outboundTag}`}
          dataSource={bindings}
          locale={{ emptyText: '尚未创建出口池绑定' }}
          columns={[
            { title: '入站', dataIndex: 'inboundTag' },
            { title: '出口', dataIndex: 'outboundTag' },
            {
              title: '操作',
              render: (_, row) => (
                <Button danger size="small" onClick={() => unbind(row.inboundTag)}>
                  解绑
                </Button>
              ),
            },
          ]}
        />
      </Card>
    </Space>
  );
}

