module.exports = {
  extract: {
    name: 'obshell',
    sourcePath: 'src',
    fileType: 'ts',
    prettier: true,
    exclude: path => {
      const excludeList = [
        '/.umi',
        '/constant/must-ignore.ts',
        '/constant/datetime.ts',
        '/global.ts',
        '/locale',
      ];
      return excludeList.some(v => {
        return path.includes(v);
      });
    },
    macro: {
      path: 'src/i18n',
      method: 'formatMessage({id:"$key$", defaultMessage:"$defaultMessage$"})',
      import: "import { formatMessage } from '@/util/intl'",
      keySplitter: '.',
      placeholder: variable => {
        return `{${variable}}`;
      },
      dependencies: ['@alipay/bailingual-sdk-glue'],
    },
    babel: {
      allowImportExportEverywhere: true,
      decoratorsBeforeExport: true,
      plugins: [
        'asyncGenerators',
        'classProperties',
        'decorators-legacy',
        'doExpressions',
        'exportExtensions',
        'exportDefaultFrom',
        'typescript',
        'functionSent',
        'functionBind',
        'jsx',
        'objectRestSpread',
        'dynamicImport',
        'numericSeparator',
        'optionalChaining',
        'optionalCatchBinding',
      ],
    },
    isNeedUploadCopyToBailingual: true,
    sourceLang: 'zh-CN',
    sdkVersion: 'normal',
    bailingual: {
      appName: 'ocp-v2',
    },
  },
  import: {
    type: 'json',
    path: 'src/locale',
    bailingual: {
      appName: 'obshell',
      tag: '',
    },
  },
};
