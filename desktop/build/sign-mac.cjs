// electron-builder 26 は afterPack が desktop/ の外だと失敗する。
// ad-hoc の実体は scripts/sign-mac.cjs（darwin 以外は何もしない）。
module.exports = require("../../scripts/sign-mac.cjs");
