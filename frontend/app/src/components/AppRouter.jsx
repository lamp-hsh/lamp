import { BrowserRouter, Route, Routes } from "react-router-dom";
import AuthProvider from "./auth/AuthContext";
import Main from "./ui/main/Main";
import Home from "./ui/Home";
import SignIn from "./auth/SignIn";
import SignOut from "./auth/SignOut";
import LoginCheck from "./auth/LoginComplete";
import Storage from "./ui/storage/Storage";
import Comic from "./ui/media/Comic";
import ComicDetail from "./ui/media/ComicDetail";
import Video from "./ui/media/Video";

const AppRouter = () => {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Main>
          <Routes>
            <Route index element={<Home />} />
            <Route path="signin" element={<SignIn />} />
            <Route path="signout" element={<SignOut />} />
            <Route path="check" element={<LoginCheck />} />
            <Route path="storage" element={<Storage />} />
            <Route path="media/video" element={<Video />} />
            <Route path="media/comic" element={<Comic />} />
            <Route path="media/comic/detail" element={<ComicDetail />} />
            {/* </Route> */}
            {/* <Route path="*" element={<NotFound />} /> */}
          </Routes>
        </Main>
      </AuthProvider>
    </BrowserRouter>
  );
};

export default AppRouter;
