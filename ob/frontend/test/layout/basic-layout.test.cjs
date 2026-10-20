const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { createRequire } = require('node:module');
const vm = require('node:vm');
const { test } = require('node:test');
const ts = require('typescript');

// React is provided by Umi. Resolve from Umi so this also works with pnpm's
// isolated dependency layout, without introducing a second React instance.
const umiRequire = createRequire(require.resolve('@umijs/max/package.json'));
const React = umiRequire('react');
const { renderToStaticMarkup } = umiRequire('react-dom/server');
const element = React.createElement;
const noop = () => {};
const Passthrough = ({ children }) => element('div', null, children);
const Outlet = () => element('span', { 'data-route': true }, 'route content');
const menu = Object.assign(Passthrough, { Item: Passthrough });
const mocks = {
  '@/hook/useMenu': { useBasicMenu: () => [] },
  '@/hook/useUiMode': { default: () => ({ isDesktopMode: false }), __esModule: true },
  '@/service/obshell/task': { getAgentUnfinishDags: noop, getClusterUnfinishDags: noop },
  '@/service/obshell/v1': { getAgentInfo: noop, getTime: noop },
  '@/util/intl': { formatMessage: ({ defaultMessage }) => defaultMessage },
  '@oceanbase/design': {
    Alert: Passthrough,
    Badge: Passthrough,
    Menu: menu,
    Modal: { confirm: noop },
    Space: Passthrough,
    theme: { useToken: () => ({ token: {} }) },
    Tooltip: Passthrough,
  },
  '@oceanbase/icons': { LoadingOutlined: Passthrough, UnorderedListOutlined: Passthrough },
  '@oceanbase/ui': { BasicLayout: Passthrough },
  '@umijs/max': {
    history: { push: noop },
    Outlet,
    useDispatch: () => noop,
    useLocation: () => ({ pathname: '/cluster/tenant/dashbiz/user' }),
    useModel: () => ({ getObclusterInfo: noop }),
  },
  ahooks: { useRequest: () => ({ refresh: noop, loading: false }) },
  react: React,
  './index.less': {},
};

// Exercise the actual layout's render behavior, not a copy of its expression.
// Mock only providers, network hooks and the surrounding design-system shell.
const sourcePath = path.resolve(__dirname, '../../src/pages/Layout/BasicLayout/index.tsx');
const compiled = ts.transpileModule(fs.readFileSync(sourcePath, 'utf8'), {
  compilerOptions: {
    target: ts.ScriptTarget.ES2020,
    module: ts.ModuleKind.CommonJS,
    jsx: ts.JsxEmit.React,
    esModuleInterop: true,
  },
});
const moduleResult = { exports: {} };
vm.runInNewContext(compiled.outputText, {
  exports: moduleResult.exports,
  module: moduleResult,
  require: name => (Object.hasOwn(mocks, name) ? mocks[name] : require(name)),
  setTimeout,
});
const BasicLayout = moduleResult.exports.default;

test('keeps the tenant route and its password modal together', () => {
  const html = renderToStaticMarkup(
    element(BasicLayout, null, [
      element(Outlet, { key: 'route' }),
      element('div', { key: 'password', role: 'dialog' }, 'tenant password'),
    ])
  );
  assert.equal((html.match(/data-route/g) || []).length, 1);
  assert.match(html, /role="dialog"/);
  assert.match(html, /tenant password/);
});

test('renders the route when used as a Umi layout without children', () => {
  const html = renderToStaticMarkup(element(BasicLayout));
  assert.equal((html.match(/data-route/g) || []).length, 1);
});

test('does not replace explicitly empty children with a route', () => {
  assert.doesNotMatch(renderToStaticMarkup(element(BasicLayout, null, false)), /data-route/);
});
