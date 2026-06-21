module.exports = {
  devServer: {
    disableHostCheck: true,
  },
  // Asset base path. Configurable via PUBLIC_PATH so the Docker image can be
  // served at root ('/') behind the nginx reverse proxy.
  publicPath: process.env.PUBLIC_PATH || '/'
};