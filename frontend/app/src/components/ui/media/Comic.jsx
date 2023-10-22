import React, { useContext, useEffect, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { AuthContext } from "../../auth/AuthContext";
import { GetRequest } from "../../Request";
import ComicUpdate from "./ComicUpdate";
import Paging from "./Paging";
import TagSearch from "./TagSearch";

const Comic = () => {
  const { isSigned } = useContext(AuthContext);
  const [isLoaded, setIsLoaded] = useState(false);
  const [comicInfo, setComicInfo] = useState([]);
  const [render, setRender] = useState(null);
  const [totalItem, setTotalItem] = useState();
  const [page, setPage] = useState(1);
  const [tag, setTag] = useState("");
  const [typing, setTyping] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();

  const reRender = () => {
    setRender((prev) => !prev);
  };

  useEffect(() => {
    if (isSigned.status === false) {
      navigate(-1);
    }
    const urlParams = new URLSearchParams(window.location.search);

    if (urlParams.get("page") != null) {
      setPage(parseInt(urlParams.get("page")));
    } else setPage(1);
    if (urlParams.get("tag") != null) {
      setTag(urlParams.get("tag"));
    }
    (async () => {
      const comicList = await GetRequest(
        urlParams.toString() === ""
          ? "/api/media/comic"
          : `/api/media/comic?${urlParams.toString()}`
      );
      if (comicList && comicList.data) {
        setComicInfo(comicList.data.ComicInfo);
        setTotalItem(comicList.data.TotalCount);
      } else {
        setComicInfo([]);
        setTotalItem(0);
      }
      setIsLoaded(true);
    })();
  }, [isSigned, navigate, render, location.search]);

  return (
    <div>
      <div className="flex flex-wrap border-b border-stone-700/50 m-2 items-center justify-between">
        <h1 className="font-black text-stone-700 text-xl align-middle">
          Media - Comic
        </h1>
        <div className="flex">
          <TagSearch
            typing={typing}
            setTyping={setTyping}
            tag={tag}
            setTag={setTag}
          />
          <ComicUpdate action="refresh" render={reRender} />
        </div>
      </div>

      {isLoaded ? (
        comicInfo.length > 0 ? (
          <div>
            <div className="flex flex-wrap m-2 bg-stone-100 border border-stone-200">
              {comicInfo.map((comic, i) => {
                return (
                  <div
                    key={i}
                    onClick={() => {
                      navigate("/media/comic/detail", {
                        state: { comicName: comic.Title },
                      });
                    }}
                    className="w-1/3 sm:w-1/4 md:w-1/5 lg:w-1/6 xl:w-1/6 p-2 duration-300 transform hover:scale-105 flex"
                  >
                    <div className="relative transition-transform">
                      <img
                        src={comic.PrevImgPath}
                        alt=""
                        className="h-full object-cover"
                      />
                      <div className="absolute top-0 left-0 right-0 p-2 font-bold text-sm bg-black bg-opacity-75 text-white truncate">
                        <div className="flex items-center justify-between">
                          {comic.Title}
                        </div>
                      </div>
                      <div className="absolute bottom-0 left-0 right-0 p-2 bg-black bg-opacity-75 text-white truncate">
                        <div className="flex items-center justify-between">
                          <div className="text-sm">{comic.Tag}</div>
                        </div>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
            <Paging
              page={parseInt(page)}
              totalItem={totalItem}
              render={reRender}
            />
          </div>
        ) : (
          <div>Not Found Result...</div>
        )
      ) : (
        <div>Loading...</div>
      )}
    </div>
  );
};

export default Comic;
