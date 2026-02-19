export const manifest = (() => {
function __memo(fn) {
	let value;
	return () => value ??= (value = fn());
}

return {
	appDir: "_app",
	appPath: "_app",
	assets: new Set([]),
	mimeTypes: {},
	_: {
		client: {start:"_app/immutable/entry/start.BaWaM_sv.js",app:"_app/immutable/entry/app.BJcEnXeI.js",imports:["_app/immutable/entry/start.BaWaM_sv.js","_app/immutable/chunks/BjUChR8d.js","_app/immutable/chunks/uMvGNBtf.js","_app/immutable/chunks/BUApaBEI.js","_app/immutable/chunks/DBD2Xsyc.js","_app/immutable/entry/app.BJcEnXeI.js","_app/immutable/chunks/uMvGNBtf.js","_app/immutable/chunks/CzMyd91B.js","_app/immutable/chunks/CfPFij4U.js","_app/immutable/chunks/DBD2Xsyc.js","_app/immutable/chunks/8sHoEGoU.js","_app/immutable/chunks/Bdi-rYIY.js","_app/immutable/chunks/D2e79A1n.js"],stylesheets:[],fonts:[],uses_env_dynamic_public:false},
		nodes: [
			__memo(() => import('./nodes/0.js')),
			__memo(() => import('./nodes/1.js')),
			__memo(() => import('./nodes/2.js')),
			__memo(() => import('./nodes/3.js')),
			__memo(() => import('./nodes/4.js'))
		],
		remotes: {
			
		},
		routes: [
			{
				id: "/",
				pattern: /^\/$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 2 },
				endpoint: null
			},
			{
				id: "/results/[sessionId]",
				pattern: /^\/results\/([^/]+?)\/?$/,
				params: [{"name":"sessionId","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,], errors: [1,], leaf: 3 },
				endpoint: null
			},
			{
				id: "/session/[trackId]",
				pattern: /^\/session\/([^/]+?)\/?$/,
				params: [{"name":"trackId","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,], errors: [1,], leaf: 4 },
				endpoint: null
			}
		],
		prerendered_routes: new Set([]),
		matchers: async () => {
			
			return {  };
		},
		server_assets: {}
	}
}
})();

export const prerendered = new Set([]);

export const base = "";