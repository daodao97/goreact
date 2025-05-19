// 导出组件
export * from './components';

// 导出布局组件
export * from './layout';

// 导出区块组件
export * from './blocks';

// 导出工具库
export * from './lib';

// 默认导出所有内容
import * as components from './components';
import * as layout from './layout';
import * as blocks from './blocks';
import * as lib from './lib';

export default {
    components,
    layout,
    blocks,
    lib
}; 